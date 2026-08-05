cat > entrypoint.sh << 'EOF'
#!/bin/sh
echo "Running migrations..."
/bin/migrate -path /migrations -database "$DATABASE_URL" up
echo "Starting server..."
exec /bin/server
EOF
chmod +x entrypoint.sh