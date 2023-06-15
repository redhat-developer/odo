FROM registry.access.redhat.com/ubi8/nodejs-16:latest AS website-builder

RUN npm install --global yarn

RUN mkdir -p /tmp/website
WORKDIR /tmp/website

COPY . .

RUN yarn && \
    npx docusaurus build

FROM registry.access.redhat.com/ubi8/httpd-24:latest

COPY --from=website-builder /tmp/website/build /var/www/html
