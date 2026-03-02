declare global {
  interface Window {
    CONFIG: {
      API_URL: string;
    };
  }
}

export const getBackendUrl = () => {
  return window.CONFIG?.API_URL || 'http://localhost:8080';
};
