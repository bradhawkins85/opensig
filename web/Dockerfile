# ---------- build ----------
FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json* pnpm-lock.yaml* yarn.lock* ./
RUN npm install --no-audit --no-fund
COPY . .
RUN npm run build

# ---------- runtime ----------
FROM node:20-alpine AS runtime
ENV NODE_ENV=production
WORKDIR /app
COPY --from=build /app .
EXPOSE 3000
CMD ["npm","start"]
