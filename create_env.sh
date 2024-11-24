#!/bin/bash

cat > .env << 'EOL'
TURSO_DATABASE_URL=libsql://rep-up-emday4prez.turso.io
TURSO_AUTH_TOKEN=eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE3MzI0MTQxMjksImlkIjoiMTU2OTFiYWQtODUwNC00MzAxLWFjMTItZTY3MDhlYTY0MjE5In0._zXEpJ7V6XIJOL-CW9qEvCVttabNpIlHRlNnfJsjrecTXjj-YTpmXoG2rkIrPgnhZ-fL-pAsEf1SvkTYEEycDQ
JWT_SECRET=AEW1nYWTM5d5MXHLiDzhZ3sJL8QyV8LQgX00IOzgYBFfocLeOTCidIgcZDt8b9Gd
EOL

# Make sure there are no Windows line endings
dos2unix .env 2>/dev/null || true

# Print file contents for verification
echo "Created .env file with contents:"
cat -A .env  # Shows special characters

# Print directory contents
echo -e "\nDirectory contents:"
ls -la .env