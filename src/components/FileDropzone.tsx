"use client";
import {useCallback} from "react";
import {useDropzone} from "react-dropzone";

export default function FileDropzone({onFiles}:{onFiles:(files:File[])=>void}) {
  const onDrop = useCallback((accepted: File[]) => onFiles(accepted), [onFiles]);
  const {getRootProps, getInputProps, isDragActive} = useDropzone({onDrop, multiple:false});
  return (
    <div {...getRootProps()} className={`border-2 border-dashed rounded-xl p-6 bg-white text-center ${isDragActive?"border-black":"border-gray-300"}`}>
      <input {...getInputProps()} />
      <p className="text-gray-600">Drag & drop STL / OBJ / 3MF / G-code here, or click to browse.</p>
    </div>
  );
}
