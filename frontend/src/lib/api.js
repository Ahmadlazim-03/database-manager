import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

class ApiClient {
    constructor() {
        this.client = axios.create({
            baseURL: API_BASE_URL,
            headers: {
                'Content-Type': 'application/json',
            },
        });

        // Add token to requests
        this.client.interceptors.request.use((config) => {
            const token = localStorage.getItem('token');
            if (token) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        });

        // Handle auth errors
        this.client.interceptors.response.use(
            (response) => response,
            (error) => {
                if (error.response?.status === 401) {
                    localStorage.removeItem('token');
                    localStorage.removeItem('user');
                    window.location.href = '/login';
                }
                return Promise.reject(error);
            }
        );
    }

    // Auth methods
    async register(userData) {
        const response = await this.client.post('/auth/register', userData);
        return response.data;
    }

    async login(email, password) {
        const response = await this.client.post('/auth/login', { email, password });
        return response.data;
    }

    async getProfile() {
        const response = await this.client.get('/auth/profile');
        return response.data;
    }

    // Database methods
    async testConnection(connectionData) {
        const response = await this.client.post('/database/test', connectionData);
        return response.data;
    }

    async createConnection(connectionData) {
        const response = await this.client.post('/database', connectionData);
        return response.data;
    }

    async getConnections() {
        const response = await this.client.get('/database');
        return response.data;
    }

    async getDatabaseInfo(id) {
        const response = await this.client.get(`/database/${id}/info`);
        return response.data;
    }

    async deleteConnection(id) {
        const response = await this.client.delete(`/database/${id}`);
        return response.data;
    }

    // API management methods
    async createAPIKey(data) {
        const response = await this.client.post('/api-management/keys', data);
        return response.data;
    }

    async getAPIKeys() {
        const response = await this.client.get('/api-management/keys');
        return response.data;
    }

    async toggleAPIKey(id) {
        const response = await this.client.put(`/api-management/keys/${id}/toggle`);
        return response.data;
    }

    async deleteAPIKey(id) {
        const response = await this.client.delete(`/api-management/keys/${id}`);
        return response.data;
    }

    async createEndpoint(data) {
        const response = await this.client.post('/api-management/endpoints', data);
        return response.data;
    }

    async getEndpoints(databaseId = '') {
        const response = await this.client.get(`/api-management/endpoints?database_id=${databaseId}`);
        return response.data;
    }

    async toggleEndpoint(id) {
        const response = await this.client.put(`/api-management/endpoints/${id}/toggle`);
        return response.data;
    }

    async deleteEndpoint(id) {
        const response = await this.client.delete(`/api-management/endpoints/${id}`);
        return response.data;
    }

    async getLogs() {
        const response = await this.client.get('/api-management/logs');
        return response.data;
    }

    async clearLogs() {
        const response = await this.client.delete('/api-management/logs');
        return response.data;
    }

    // Database sharing methods
    async createInvitation(data) {
        const response = await this.client.post('/sharing/invitations', data);
        return response.data;
    }

    async getDatabaseInvitations(databaseId) {
        const response = await this.client.get(`/sharing/invitations/database/${databaseId}`);
        return response.data;
    }

    async getInvitation(token) {
        const response = await this.client.get(`/sharing/invitations/${token}`);
        return response.data;
    }

    async acceptInvitation(token) {
        const response = await this.client.post(`/sharing/invitations/${token}/accept`);
        return response.data;
    }

    async getSharedDatabases() {
        const response = await this.client.get('/sharing/shared-databases');
        return response.data;
    }

    async getPendingInvitations() {
        const response = await this.client.get('/sharing/pending-invitations');
        return response.data;
    }

    async getDatabaseAccess(databaseId) {
        const response = await this.client.get(`/sharing/database-access/${databaseId}`);
        return response.data;
    }

    async revokeAccess(databaseId, userId) {
        const response = await this.client.delete('/sharing/access', {
            data: { database_id: databaseId, user_id: userId }
        });
        return response.data;
    }

    async revokeInvitation(invitationId) {
        const response = await this.client.delete(`/sharing/invitations/${invitationId}`);
        return response.data;
    }

    async leaveSharedDatabase(databaseId) {
        const response = await this.client.delete('/sharing/leave', {
            data: { database_id: databaseId }
        });
        return response.data;
    }
}

export const apiClient = new ApiClient();
