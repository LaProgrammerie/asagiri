import { createMDX } from 'fumadocs-mdx/next';

const isGithubPages = process.env.GITHUB_PAGES === 'true';
const basePath = isGithubPages ? '/asagiri' : undefined;

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  trailingSlash: true,
  reactStrictMode: true,
  images: {
    unoptimized: true,
  },
  ...(basePath !== undefined
    ? { basePath, assetPrefix: basePath }
    : {}),
};

export default withMDX(nextConfig);
