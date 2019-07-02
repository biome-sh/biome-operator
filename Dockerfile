FROM scratch

COPY biome-operator /biome-operator

ENTRYPOINT ["/biome-operator", "-logtostderr"]
