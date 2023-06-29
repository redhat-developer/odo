build-wasm:
	( \
		cd wasm/ && \
		GOOS=js GOARCH=wasm go build -o ../src/assets/devfile.wasm && \
		HASH=$$(md5sum ../src/assets/devfile.wasm | awk '{ print $$1 }') && \
		echo $${HASH} && \
		mv ../src/assets/devfile.wasm ../src/assets/devfile.$${HASH}.wasm && \
		sed -i "s/devfile\.[a-z0-9]*\.wasm/devfile\.$${HASH}.wasm/" ../src/app/app.module.ts \
	)

deploy:
	npm run deploy
