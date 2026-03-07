cd ../src
go clean 
make build
cd ./frontend
pnpm run build
systemctl --user restart openclaw-manager.service


