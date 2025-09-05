import { writable } from 'svelte/store';

// Auth store
export const user = writable(null);
export const isAuthenticated = writable(false);

// Initialize auth state from localStorage
if (typeof localStorage !== 'undefined') {
    const storedUser = localStorage.getItem('user');
    const token = localStorage.getItem('token');
    
    if (storedUser && token) {
        user.set(JSON.parse(storedUser));
        isAuthenticated.set(true);
    }
}

// Auth functions
export function setUser(userData, token) {
    user.set(userData);
    isAuthenticated.set(true);
    localStorage.setItem('user', JSON.stringify(userData));
    localStorage.setItem('token', token);
}

export function logout() {
    user.set(null);
    isAuthenticated.set(false);
    localStorage.removeItem('user');
    localStorage.removeItem('token');
}

// Database connections store
export const connections = writable([]);
export const selectedConnection = writable(null);

// API keys and endpoints store
export const apiKeys = writable([]);
export const endpoints = writable([]);
export const logs = writable([]);
