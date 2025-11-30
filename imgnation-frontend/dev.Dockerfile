FROM denoland/deno:latest

# Set the working directory inside the container
WORKDIR /app

# Copy deno.json and deno.lock (if you have one) first to leverage Docker's build cache
# This step only re-runs if deno.json or deno.lock changes
COPY deno.json ./

# Clean up any existing node_modules or Deno-managed npm caches
# This helps prevent "Text file busy" errors by ensuring a clean install
RUN rm -rf ./node_modules

# Install npm dependencies managed by Deno
# The `deno cache` command will download and cache the npm packages defined in deno.json
RUN deno cache --node-modules-dir deno.json

# Copy the rest of your application code
COPY . .

# Expose the port Vite typically runs on (default for dev server)
EXPOSE 5173

# Command to run the development server
# --allow-net: Grants network access
# --allow-read: Grants read access to files
# --node-modules-dir: Ensures Deno uses the node_modules directory
CMD ["deno", "run", "-A", "--node-modules-dir", "npm:vite"]
