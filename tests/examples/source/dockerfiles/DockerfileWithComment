# Install the app dependencies in a full Node docker image
FROM node:12

COPY . /project/

# Removing node_modules
RUN rm -rf /project/node_modules

# Install user-app dependencies
WORKDIR /project
RUN npm install --production 

# Copy the dependencies into a slim Node docker image
FROM node:12-slim

# Copy project with dependencies
COPY --chown=node:node --from=0 /project /project

WORKDIR /project

ENV NODE_ENV production

USER node

CMD ["npm", "start"]