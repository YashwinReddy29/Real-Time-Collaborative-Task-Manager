import axios from "axios";

export function getBackendOrigin(): string {
  const configured = import.meta.env.VITE_API_ORIGIN?.trim();

  return configured
    ? configured.replace(/\/+$/, "")
    : window.location.origin;
}

const api = axios.create({
  baseURL: `${getBackendOrigin()}/api`,
  timeout: 15000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (
      axios.isAxiosError(error) &&
      error.response?.status === 401
    ) {
      localStorage.removeItem("token");

      if (window.location.pathname !== "/login") {
        window.location.assign("/login");
      }
    }

    return Promise.reject(error);
  }
);

export default api;
