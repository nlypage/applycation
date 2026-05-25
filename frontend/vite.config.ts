import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const frontendPort = Number(process.env.FRONTEND_PORT || 5173);
const apiBaseURL = process.env.VITE_API_BASE_URL || "http://localhost:8080";

export default defineConfig({
	plugins: [react()],
	server: {
		host: "0.0.0.0",
		port: frontendPort,
		proxy: {
			"/health": {
				target: apiBaseURL,
				changeOrigin: true,
			},
		},
	},
});
