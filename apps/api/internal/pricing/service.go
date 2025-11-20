package pricing

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Options struct {
	MaterialCostPLA float64
	MachineRate     float64
	SetupFee        float64
	PrintSpeed      float64
	Logger          *slog.Logger
	StoragePath     string
}

type Service struct {
	opts Options
}

type Estimate struct {
	ID               uuid.UUID       `json:"id"`
	Material         string          `json:"material"`
	MaterialCost     float64         `json:"materialCost"`
	EstimatedGrams   float64         `json:"estimatedGrams"`
	EstimatedHours   float64         `json:"estimatedHours"`
	EstimatedPrice   float64         `json:"estimatedPrice"`
	SetupFee         float64         `json:"setupFee"`
	MachineRate      float64         `json:"machineRate"`
	PrintSpeed       float64         `json:"printSpeed"`
	Density          float64         `json:"density"`
	FileName         string          `json:"fileName"`
	FileSizeBytes    int64           `json:"fileSizeBytes"`
	TriangleCount    int             `json:"triangleCount"`
	BoundingBoxMM    BoundingBox     `json:"boundingBoxMm"`
	VolumeCM3        float64         `json:"volumeCm3"`
	SurfaceAreaCM2   float64         `json:"surfaceAreaCm2"`
	Confidence       string          `json:"confidence"`
	Warnings         []string        `json:"warnings"`
	Metadata         map[string]any  `json:"metadata"`
	RecommendedInfill int            `json:"recommendedInfill"`
}

type BoundingBox struct {
	Min [3]float64 `json:"min"`
	Max [3]float64 `json:"max"`
}

func NewService(opts Options) *Service {
	return &Service{opts: opts}
}

func (s *Service) EstimateFromUpload(ctx context.Context, fileHeader *multipart.FileHeader) (*Estimate, []byte, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, nil, err
	}
	defer src.Close()
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, src); err != nil {
		return nil, nil, err
	}
	data := buf.Bytes()

	analysis, warn := s.analyseGeometry(fileHeader.Filename, data)
	estimate := s.pricingFor(analysis)
	estimate.FileName = fileHeader.Filename
	estimate.FileSizeBytes = int64(len(data))
	estimate.Warnings = warn
	estimate.Metadata = map[string]any{
		"generatedAt": time.Now().UTC(),
	}
	return estimate, data, nil
}

type geometry struct {
	TriangleCount int
	BoundingBox   BoundingBox
	VolumeCM3     float64
	SurfaceArea   float64
	Confidence    string
}

func (s *Service) pricingFor(g geometry) *Estimate {
	const densityPLA = 1.24 // g/cm3
	grams := math.Max(8, g.VolumeCM3*densityPLA*1.05) // add 5% margin
	if g.VolumeCM3 == 0 {
		// fallback heuristic using bounding box diagonal
		size := diagonal(g.BoundingBox)
		grams = math.Max(8, size*0.9)
	}
	hours := grams / 12.0
	if s.opts.PrintSpeed > 0 && g.VolumeCM3 > 0 {
		printHours := (g.VolumeCM3 * 1000) / s.opts.PrintSpeed
		hours = math.Max(hours, printHours)
	}
	setup := s.opts.SetupFee
	machine := hours * s.opts.MachineRate
	material := grams * s.opts.MaterialCostPLA
	total := setup + machine + material
	infill := 20
	if grams > 120 {
		infill = 15
	}
	if grams < 30 {
		infill = 25
	}
	return &Estimate{
		ID:               uuid.New(),
		Material:         "PLA",
		MaterialCost:     material,
		EstimatedGrams:   round1(grams),
		EstimatedHours:   round2(hours),
		EstimatedPrice:   round2(total),
		SetupFee:         setup,
		MachineRate:      s.opts.MachineRate,
		PrintSpeed:       s.opts.PrintSpeed,
		Density:          densityPLA,
		TriangleCount:    g.TriangleCount,
		BoundingBoxMM:    g.BoundingBox,
		VolumeCM3:        round2(g.VolumeCM3),
		SurfaceAreaCM2:   round2(g.SurfaceArea),
		Confidence:       g.Confidence,
		RecommendedInfill: infill,
	}
}

func (s *Service) analyseGeometry(name string, data []byte) (geometry, []string) {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".stl":
		if g, err := parseBinarySTL(data); err == nil {
			return g, nil
		}
	case ".obj":
		if g, err := parseOBJ(data); err == nil {
			return g, nil
		}
	case ".3mf":
		if g, err := parse3MF(data); err == nil {
			return g, nil
		}
	}
	// fallback heuristic based on size
	size := len(data)
	grams := math.Max(8, math.Min(250, float64(size)/7000))
	bb := BoundingBox{
		Min: [3]float64{0, 0, 0},
		Max: [3]float64{grams * 0.9, grams * 0.5, grams * 0.4},
	}
	return geometry{
		TriangleCount: size / 50,
		BoundingBox:   bb,
		VolumeCM3:     grams / 1.24,
		SurfaceArea:   grams * 1.5,
		Confidence:    "low",
	}, []string{"Used heuristic estimation because detailed geometry parsing failed."}
}

func parseBinarySTL(data []byte) (geometry, error) {
	if len(data) < 84 {
		return geometry{}, errors.New("stl too small")
	}
	triCount := int(binary.LittleEndian.Uint32(data[80:84]))
	offset := 84
	if len(data) < offset+50*triCount {
		// likely ASCII STL; fallback
		return parseASCIISTL(data)
	}
	var (
		minX, minY, minZ = math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
		maxX, maxY, maxZ = -math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
		volume           float64
		area             float64
	)

	for i := 0; i < triCount; i++ {
		if offset+50 > len(data) {
			break
		}
		// skip normal (12 bytes)
		offset += 12
		v := make([][3]float64, 3)
		for j := 0; j < 3; j++ {
			x := math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
			y := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
			z := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+8 : offset+12]))
			offset += 12
			v[j] = [3]float64{float64(x), float64(y), float64(z)}
			minX = math.Min(minX, float64(x))
			minY = math.Min(minY, float64(y))
			minZ = math.Min(minZ, float64(z))
			maxX = math.Max(maxX, float64(x))
			maxY = math.Max(maxY, float64(y))
			maxZ = math.Max(maxZ, float64(z))
		}
		volume += signedVolumeOfTriangle(v[0], v[1], v[2])
		area += triangleArea(v[0], v[1], v[2])
		offset += 2 // attribute byte count
	}
	if triCount == 0 {
		return geometry{}, errors.New("no triangles parsed")
	}
	bb := BoundingBox{
		Min: [3]float64{minX, minY, minZ},
		Max: [3]float64{maxX, maxY, maxZ},
	}
	return geometry{
		TriangleCount: triCount,
		BoundingBox:   bb,
		VolumeCM3:     math.Abs(volume) / 1000.0,
		SurfaceArea:   area / 100.0,
		Confidence:    "high",
	}, nil
}

func parseASCIISTL(data []byte) (geometry, error) {
	lines := bytes.Split(data, []byte("\n"))
	var (
		minX, minY, minZ = math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
		maxX, maxY, maxZ = -math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
		volume           float64
		area             float64
		triCount         int
		current          [][3]float64
	)
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if bytes.HasPrefix(line, []byte("vertex")) {
			var x, y, z float64
			if _, err := fmtSscanf(string(line), "vertex %f %f %f", &x, &y, &z); err == nil {
				current = append(current, [3]float64{x, y, z})
				minX = math.Min(minX, x)
				minY = math.Min(minY, y)
				minZ = math.Min(minZ, z)
				maxX = math.Max(maxX, x)
				maxY = math.Max(maxY, y)
				maxZ = math.Max(maxZ, z)
			}
		}
		if bytes.HasPrefix(line, []byte("endfacet")) {
			if len(current) >= 3 {
				v0, v1, v2 := current[0], current[1], current[2]
				volume += signedVolumeOfTriangle(v0, v1, v2)
				area += triangleArea(v0, v1, v2)
				triCount++
			}
			current = current[:0]
		}
	}
	if triCount == 0 {
		return geometry{}, errors.New("ascii stl parse failed")
	}
	bb := BoundingBox{
		Min: [3]float64{minX, minY, minZ},
		Max: [3]float64{maxX, maxY, maxZ},
	}
	return geometry{
		TriangleCount: triCount,
		BoundingBox:   bb,
		VolumeCM3:     math.Abs(volume) / 1000.0,
		SurfaceArea:   area / 100.0,
		Confidence:    "medium",
	}, nil
}

func parseOBJ(data []byte) (geometry, error) {
	lines := bytes.Split(data, []byte("\n"))
	var vertices [][3]float64
	var faces [][3]int
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if bytes.HasPrefix(line, []byte("v ")) {
			var x, y, z float64
			if _, err := fmtSscanf(string(line), "v %f %f %f", &x, &y, &z); err == nil {
				vertices = append(vertices, [3]float64{x, y, z})
			}
		}
		if bytes.HasPrefix(line, []byte("f ")) {
			var a, b, c int
			if _, err := fmtSscanf(string(line), "f %d %d %d", &a, &b, &c); err == nil {
				faces = append(faces, [3]int{a - 1, b - 1, c - 1})
			}
		}
	}
	if len(vertices) == 0 || len(faces) == 0 {
		return geometry{}, errors.New("obj missing vertices/faces")
	}
	minX, minY, minZ := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	maxX, maxY, maxZ := -math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, v := range vertices {
		minX = math.Min(minX, v[0])
		minY = math.Min(minY, v[1])
		minZ = math.Min(minZ, v[2])
		maxX = math.Max(maxX, v[0])
		maxY = math.Max(maxY, v[1])
		maxZ = math.Max(maxZ, v[2])
	}
	var volume, area float64
	for _, f := range faces {
		v0 := vertices[f[0]]
		v1 := vertices[f[1]]
		v2 := vertices[f[2]]
		volume += signedVolumeOfTriangle(v0, v1, v2)
		area += triangleArea(v0, v1, v2)
	}
	bb := BoundingBox{
		Min: [3]float64{minX, minY, minZ},
		Max: [3]float64{maxX, maxY, maxZ},
	}
	return geometry{
		TriangleCount: len(faces),
		BoundingBox:   bb,
		VolumeCM3:     math.Abs(volume) / 1000.0,
		SurfaceArea:   area / 100.0,
		Confidence:    "medium",
	}, nil
}

func parse3MF(data []byte) (geometry, error) {
	// TODO: implement proper 3MF parsing. For now return heuristic error.
	return geometry{}, errors.New("3mf parsing not implemented")
}

func signedVolumeOfTriangle(p1, p2, p3 [3]float64) float64 {
	return (p1[0]*p2[1]*p3[2] + p2[0]*p3[1]*p1[2] + p3[0]*p1[1]*p2[2] - p1[0]*p3[1]*p2[2] - p2[0]*p1[1]*p3[2] - p3[0]*p2[1]*p1[2]) / 6.0
}

func triangleArea(p1, p2, p3 [3]float64) float64 {
	ax := p2[0] - p1[0]
	ay := p2[1] - p1[1]
	az := p2[2] - p1[2]
	bx := p3[0] - p1[0]
	by := p3[1] - p1[1]
	bz := p3[2] - p1[2]
	cx := ay*bz - az*by
	cy := az*bx - ax*bz
	cz := ax*by - ay*bx
	return 0.5 * math.Sqrt(cx*cx+cy*cy+cz*cz)
}

func diagonal(bb BoundingBox) float64 {
	dx := bb.Max[0] - bb.Min[0]
	dy := bb.Max[1] - bb.Min[1]
	dz := bb.Max[2] - bb.Min[2]
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// fmtSscanf wraps fmt.Sscanf but avoids importing fmt globally in this file.
func fmtSscanf(str, format string, a ...any) (int, error) {
	return fmt.Sscanf(str, format, a...)
}
