# Official Playwright image — includes Chromium, Firefox, WebKit + all OS deps.
# Pinned to match the @playwright/test version in package.json.
FROM mcr.microsoft.com/playwright:v1.50.0-noble

WORKDIR /tests

COPY package.json ./
RUN npm install --save-exact

COPY . .

CMD ["npx", "playwright", "test", "--reporter=list"]
