# Use the official Go image as base
FROM golang:1.25

# Install nano text editor, curl, and Node.js (required for Claude Code)
RUN apt-get update && apt-get install -y \
    nano \
    curl \
    sqlite3 \
    libsqlite3-dev \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/* 

# Set the working directory inside the container
WORKDIR /app

# Set bash as the default shell
CMD ["/bin/bash"]
