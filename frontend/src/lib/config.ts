declare global {
  interface Window {
    CONFIG: {
      BACKEND_URL: string;
    };
  }
}

export const getBackendUrl = () => {
  return window.CONFIG?.BACKEND_URL || 'http://localhost:8080';
};
