// Environment configuration helper
export const config = {
    // API Configuration
    API_BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
    BACKEND_PORT: import.meta.env.VITE_BACKEND_PORT || '8080',
    
    // Helper function to get full API URL
    getApiUrl: (endpoint = '') => {
        const baseUrl = import.meta.env.VITE_API_BASE_URL || `http://localhost:${import.meta.env.VITE_BACKEND_PORT || '8080'}/api`;
        return endpoint ? `${baseUrl}${endpoint.startsWith('/') ? '' : '/'}${endpoint}` : baseUrl;
    },
    
    // Helper function to get backend base URL (without /api)
    getBackendUrl: () => {
        const port = import.meta.env.VITE_BACKEND_PORT || '8080';
        return `http://localhost:${port}`;
    }
};

export default config;
