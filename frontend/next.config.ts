import type { NextConfig } from "next";

const nextConfig: NextConfig = {
    /* config options here */
    output: "export",
    trailingSlash: true,
    turbopack: {
        rules: {
            "*.css": {
                loaders: ["@tailwindcss/turbopack"],
                as: "*.css",
            },
        },
    },
    async rewrites() {
        return [
            {
                source: "/api/:path*",
                destination: `http://localhost:8001/api/:path*`,
            },
            {
                source: "/auth/:path*",
                destination: `http://localhost:8001/auth/:path*`,
            },
        ]
    }
};

export default nextConfig;
