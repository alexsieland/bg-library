declare global {
  interface Window {
    CONFIG: {
      API_URL: string;
      BARCODE_ENABLED: boolean;
    };
  }
}

export const getBackendUrl = () => {
  return window.CONFIG?.API_URL || "http://localhost:8080";
};

export const isBarcodeEnabled = () => {
  return window.CONFIG?.BARCODE_ENABLED || false;
};
