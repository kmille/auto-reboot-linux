build:
	makepkg --syncdeps
install:
	sudo pacman -U *.tar.zst
buildInstall:
	makepkg --syncdeps --install --force
clean:
	rm -rf src pkg *.zst keys *.xz *.asc
