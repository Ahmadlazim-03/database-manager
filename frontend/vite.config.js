import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 3000,
		proxy: {
			'/api': {
				target: 'http://localhost:8081',
				changeOrigin: true,
				bypass: function (req, res, proxyOptions) {
					// Don't proxy frontend routes like /api-management page
					if (req.url.startsWith('/api-management') && !req.url.startsWith('/api/api-management')) {
						return req.url;
					}
				}
			}
		}
	}
});
