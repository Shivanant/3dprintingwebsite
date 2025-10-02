export function quickEstimate(bytes: number, material = "PLA", quality = "standard") {
  const grams = Math.max(5, Math.min(300, bytes / 10000));      // quick proxy
  const matPerGram = material === "PLA" ? 1.2 : 1.6;            // ₹/g example
  const timeHours = Math.max(0.5, grams / 20);                  // proxy
  const machineRate = quality === "fine" ? 200 : 150;           // ₹/hr
  const cost = grams*matPerGram + timeHours*machineRate + 50;   // setup fee
  return { grams: +grams.toFixed(1), timeHours:+timeHours.toFixed(2), total: Math.round(cost) };
}
