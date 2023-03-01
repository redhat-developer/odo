# Copied base image from docker hub to quay.io  
# skopeo copy --all docker://node:19 docker://quay.io/odo-dev/node:19

FROM quay.io/odo-dev/node:19

WORKDIR /usr/src/app

COPY package*.json ./
RUN npm install
COPY . .

EXPOSE 8080
CMD [ "node", "server.js" ]
