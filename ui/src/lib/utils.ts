import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatBytes(bytes: number, decimals = 2) {
  if (bytes === 0) return "0 Bytes";
  
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB"];
  
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}

export function formatDate(date: string | Date) {
  return new Date(date).toLocaleString();
}

export function getStatusColor(status: string) {
  switch (status.toLowerCase()) {
    case "running":
      return "bg-green-500";
    case "stopped":
      return "bg-red-500";
    case "paused":
      return "bg-yellow-500";
    case "shutdown":
      return "bg-gray-500";
    case "crashed":
      return "bg-red-700";
    case "unknown":
    default:
      return "bg-gray-400";
  }
}

export function getStatusTextColor(status: string) {
  switch (status.toLowerCase()) {
    case "running":
      return "text-green-500";
    case "stopped":
      return "text-red-500";
    case "paused":
      return "text-yellow-500";
    case "shutdown":
      return "text-gray-500";
    case "crashed":
      return "text-red-700";
    case "unknown":
    default:
      return "text-gray-400";
  }
}

export function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}