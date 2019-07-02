pkg_name="biome-operator"
pkg_origin="biome"
pkg_version=$(cat "$PLAN_CONTEXT/../VERSION")
pkg_description="A Kubernetes operator for Biome services"
pkg_upstream_url="https://github.com/biome-sh/biome-operator"
pkg_license=('Apache-2.0')
pkg_maintainer="The Biome Maintainers <humans@biome.sh>"
pkg_bin_dirs=(bin)
scaffolding_go_base_path=github.com/biome-sh
pkg_scaffolding=core/scaffolding-go
pkg_svc_run="${pkg_name}"

do_build() {
  pushd "$scaffolding_go_pkg_path" >/dev/null
  make -e BIN_PATH="${scaffolding_go_gopath}/bin/biome-operator" linux
  popd >/dev/null
}

do_install() {
  cp -r "${scaffolding_go_gopath}/bin" "${pkg_prefix}/${bin}"
}
