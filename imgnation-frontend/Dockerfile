# --- Build Stage ---
FROM denoland/deno:latest as builder

WORKDIR /app

# Copy deno.json and deno.lock (if you have one)
COPY deno.json ./

# Install npm dependencies
RUN deno cache --node-modules-dir deno.json

# Copy the rest of your source code
COPY . .

# Build the React application using the 'build' task defined in deno.json
# This will create the 'dist' directory with the optimized production build
RUN deno task build

# --- Production Stage ---
# Use a smaller Deno image for the final production environment
FROM denoland/deno:1.44.4-alpine

WORKDIR /app

# Copy only the built application from the build stage
COPY --from=builder /app/dist ./dist

# Copy deno.json to access the 'serve' task
COPY deno.json ./deno.json

# Expose the port that the Deno file server will listen on (as defined in deno.json 'serve' task)
EXPOSE 3000

# Command to serve the static files using the 'serve' task from deno.json
# --allow-net: Grants network access
# --allow-read: Grants read access to the 'dist' directory
CMD ["deno", "task", "serve"]
