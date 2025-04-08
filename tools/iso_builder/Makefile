.PHONY: cleanall clean-appliance-temp-dir build-ove-iso
ARCH ?= x86_64

cleanall:
	@echo "🧹 Cleaning directory: /tmp/iso_builder";
	@rm -rf /tmp/iso_builder; 
	@echo "✅ Done."

clean-appliance-temp-dir:
	@read -p "🧹 Enter the appliance directory path to clean (e.g., /tmp/iso_builder/<OCP_VERSION>/appliance): " dir; \
	if [ -d "$$dir" ]; then \
		if [ -d "$$dir/temp" ]; then \
			echo "🧹 Cleaning directory: $$dir/appliance"; \
			sudo podman run --rm -it --privileged --net=host -v "$$dir/:/assets:Z" quay.io/edge-infrastructure/openshift-appliance:latest clean; \
			echo "✅ Done."; \
		else \
			echo "❌ Error: The directory '$$dir/' does not contain a 'temp' folder."; \
			echo "ℹ️ This may indicate that the appliance directory has already been cleaned, or the path is incorrect."; \
			exit 1; \
		fi \
	else \
		echo "❌ Error: Directory '$$dir' does not exist."; \
		exit 1; \
	fi
	
build-ove-iso:
	@missing=0; \
	for var in RELEASE_IMAGE_URL PULL_SECRET_FILE; do \
	  if [ -z "$${!var}" ]; then \
	    echo "❌  Error: Environment variable '$$var' is required."; \
	    missing=1; \
	  fi; \
	done; \
	if [ "$$missing" -eq 1 ]; then \
	  echo ""; \
	  echo "ℹ️   Usage: make build-ove-iso RELEASE_IMAGE_URL=<RELEASE_IMAGE> PULL_SECRET_FILE=</path/to/pull-secret.json>"; \
	  exit 1; \
	fi; \
	hack/build-ove-image.sh --release-image-url ${RELEASE_IMAGE_URL} --pull-secret-file ${PULL_SECRET_FILE}
