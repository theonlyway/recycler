# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.9.2] - 2026-01-30
### :bug: Bug Fixes
- [`f060792`](https://github.com/theonlyway/recycler/commit/f06079242eb8f5b17915a54723f450e0e8fda447) - **deps**: update go dependencies *(PR [#73](https://github.com/theonlyway/recycler/pull/73) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.9.1] - 2026-01-29
### :bug: Bug Fixes
- [`0ff48a3`](https://github.com/theonlyway/recycler/commit/0ff48a3de7092d43ccdccce2e2ff6fa8d60f99bf) - enhance controller startup diagnostics by capturing debug info on failure *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`cf4357f`](https://github.com/theonlyway/recycler/commit/cf4357ff311df2e128f4e5edd58a78f345a5f31c) - add setup for kind cluster and kubectl configuration in dev container *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.9.0] - 2026-01-29
### :sparkles: New Features
- [`c8ad927`](https://github.com/theonlyway/recycler/commit/c8ad92791493e90f5b2036f7d5326c99e4118673) - add preservation of 'latest' tag for GHCR and Docker Hub, and cleanup for failed Helm chart versions *(commit by [@theonlyway](https://github.com/theonlyway))*

### :bug: Bug Fixes
- [`a38b238`](https://github.com/theonlyway/recycler/commit/a38b2386ce27374bb2e8b1026710845bf0f1aac4) - remove failed Helm chart deletion step from workflow *(commit by [@theonlyway](https://github.com/theonlyway))*

### :recycle: Refactors
- [`92f871a`](https://github.com/theonlyway/recycler/commit/92f871aff227f4ae151a5a4f7c0d5983ca2a1ab9) - restructure build workflow by removing redundant steps and enhancing Helm test dependencies *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`14559e2`](https://github.com/theonlyway/recycler/commit/14559e298d0941e80a9e1702ca12c696f1d33007) - **deps**: update docker/login-action digest to c94ce9f *(PR [#70](https://github.com/theonlyway/recycler/pull/70) by [@renovate[bot]](https://github.com/apps/renovate))*
- [`19acb41`](https://github.com/theonlyway/recycler/commit/19acb41512f93e73ab110c13ef1dc02384604122) - **deps**: update mcr.microsoft.com/devcontainers/go:2.0-1.25-bookworm docker digest to 0624dcd *(PR [#71](https://github.com/theonlyway/recycler/pull/71) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.8.2] - 2026-01-28
### :bug: Bug Fixes
- [`d9f2e58`](https://github.com/theonlyway/recycler/commit/d9f2e5846e8316543d9e93624651c309c8d2c342) - add GitHub App token generation for changelog updates and adjust checkout step *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`6125182`](https://github.com/theonlyway/recycler/commit/61251824375e5b962ef7974502f8cd69780621dd) - **deps**: update actions/create-github-app-token action to v2 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.7.1] - 2026-01-27
### :bug: Bug Fixes
- [`5265a20`](https://github.com/theonlyway/recycler/commit/5265a2039502e87c70e16446fa812d36785db246) - log pod termination events before deletion in terminatePods function *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`a802c61`](https://github.com/theonlyway/recycler/commit/a802c61decffe33901bd6ac80f2c6ec69dac34a8) - verify target deployment existence and update Recycler status accordingly *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`be9208e`](https://github.com/theonlyway/recycler/commit/be9208e2740634936ff97c25841eecdeacf5ff76) - update event recording in terminatePods function to reference Recycler CR *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`7f825f2`](https://github.com/theonlyway/recycler/commit/7f825f2d1947b17a60cec3c9d3d59dcb0676e30c) - capture events related to Recycler CR in e2e tests *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`2486515`](https://github.com/theonlyway/recycler/commit/24865155cfb7b534483e07202eba1403ca1bcafc) - capture all events in test namespace for debugging in e2e tests *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`699ae7d`](https://github.com/theonlyway/recycler/commit/699ae7df92479ba729603734686b706f607898ad) - comment out Prometheus and cert-manager installation in e2e tests *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`3681b7f`](https://github.com/theonlyway/recycler/commit/3681b7fc346f05a5caec3558de9840f2995084e4) - update RBAC rules for events and enhance error logging in controllers *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`40df4e6`](https://github.com/theonlyway/recycler/commit/40df4e6f912247ae81ee8ef539ba6747b95f1022) - add rollout status check for metrics-server installation *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`b2aa102`](https://github.com/theonlyway/recycler/commit/b2aa1024f25b7c6aef89e7149417c663285183b4) - update metrics-server installation command to replace metric-resolution argument *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`5c28718`](https://github.com/theonlyway/recycler/commit/5c28718df8129882b2aecfefda2ef5a36c51c57d) - uncomment Prometheus operator and cert-manager installation in e2e tests *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`4a4cfff`](https://github.com/theonlyway/recycler/commit/4a4cfff727dc3e56e49068f1becbb9420be09d2d) - add debug logs for metrics-server pod status and logs retrieval *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`dd47e61`](https://github.com/theonlyway/recycler/commit/dd47e61755c7cea4fa4f612721b95e85bf32976b) - enhance metrics-server installation with detailed deployment and pod status logging *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`a94dc94`](https://github.com/theonlyway/recycler/commit/a94dc942b79f4ba86139af174217bf7bbad07bd5) - add kubelet request timeout argument to metrics-server deployment *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`c53eaf3`](https://github.com/theonlyway/recycler/commit/c53eaf36aef43bd5393c1b34f5db16182c115457) - enhance pod metrics handling and improve metrics-server patch command *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`99edf95`](https://github.com/theonlyway/recycler/commit/99edf95da59e99cdbf9dd62136f6ea51215cccf6) - update Go version in Dockerfile to 2.0-1.25-bookworm *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`618e061`](https://github.com/theonlyway/recycler/commit/618e061b58d6b582a2244433f7ce5f4f06b3088f) - update paths-ignore to exclude README.md in build workflow [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`e232f0e`](https://github.com/theonlyway/recycler/commit/e232f0e8d6db42ed5e188f41f328d005663502fc) - **deps**: pin mcr.microsoft.com/devcontainers/go docker tag to ef7d7fe *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.6.6] - 2026-01-20
### :wrench: Chores
- [`7015be7`](https://github.com/theonlyway/recycler/commit/7015be77a7e9392f173857be9d47b27c7f959e0a) - **deps**: update actions/download-artifact action to v7 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`5164c4f`](https://github.com/theonlyway/recycler/commit/5164c4f58e8648e36e029d1de60567bb9ed60f87) - **deps**: update golang:1.25 docker digest to ce63a16 *(PR [#57](https://github.com/theonlyway/recycler/pull/57) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.6.5] - 2026-01-16
### :wrench: Chores
- [`b2af9da`](https://github.com/theonlyway/recycler/commit/b2af9da06251395196dc9edd1e917d552d53b36d) - **deps**: update dependency go to v1.25.6 *(PR [#54](https://github.com/theonlyway/recycler/pull/54) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.6.4] - 2026-01-16
### :wrench: Chores
- [`638756d`](https://github.com/theonlyway/recycler/commit/638756da20d0fea135b9cafb8a590674c25d5027) - **deps**: update golang:1.25 docker digest to bc45dfd *(PR [#55](https://github.com/theonlyway/recycler/pull/55) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.6.3] - 2026-01-14
### :wrench: Chores
- [`08d3e6c`](https://github.com/theonlyway/recycler/commit/08d3e6ca5fbeb0ecf19c2d508d546f623ad3d905) - **deps**: update golang:1.25 docker digest to 8bbd140 *(PR [#53](https://github.com/theonlyway/recycler/pull/53) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.6.2] - 2026-01-13
### :wrench: Chores
- [`eafbc08`](https://github.com/theonlyway/recycler/commit/eafbc08a5534efb8daaa35f29a2f9f5489ea6267) - **deps**: update golang:1.25 docker digest to 0f406d3 *(PR [#51](https://github.com/theonlyway/recycler/pull/51) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.6.1] - 2026-01-13
### :bug: Bug Fixes
- [`f2702a0`](https://github.com/theonlyway/recycler/commit/f2702a0e698fddf6038ebbd584cc16eaef9100f0) - enhance SBOM file handling in release process *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.6.0] - 2026-01-13
### :sparkles: New Features
- [`2589fd3`](https://github.com/theonlyway/recycler/commit/2589fd3aabcd9c39cc13d4c1f8025be6dcb889b6) - add support for CycloneDX SBOM generation alongside SPDX in build workflow *(commit by [@theonlyway](https://github.com/theonlyway))*

### :bug: Bug Fixes
- [`319cd2a`](https://github.com/theonlyway/recycler/commit/319cd2a803002b7b9c1a1ca801e249b2165f2056) - **deps**: update module github.com/onsi/ginkgo/v2 to v2.27.5 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*

### :wrench: Chores
- [`510b917`](https://github.com/theonlyway/recycler/commit/510b9178635a52c4f51a694e48b626b577f6b027) - **deps**: pin dependencies *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`26edfd1`](https://github.com/theonlyway/recycler/commit/26edfd17352f3b83c249ac660881b52e412ad5d5) - **deps**: update actions/attest-build-provenance action to v3 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`fc53a96`](https://github.com/theonlyway/recycler/commit/fc53a964a1c345ef92cf962ca3241d7fd736b77b) - **deps**: update actions/setup-go digest to 7a3fe6c *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`216dba0`](https://github.com/theonlyway/recycler/commit/216dba09003aa961078a96603c23406c42794565) - **deps**: update actions/attest-sbom action to v3 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`ae4add0`](https://github.com/theonlyway/recycler/commit/ae4add00417ced96f50273fdada78cfca6f36011) - add concurrency settings to workflow files [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.5.1] - 2026-01-13
### :bug: Bug Fixes
- [`e58f4d2`](https://github.com/theonlyway/recycler/commit/e58f4d2db4cb9027b129fab3dcc8bd5675548d30) - rename Helm chart package for clarity and update references in workflow *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.5.0] - 2026-01-13
### :sparkles: New Features
- [`e880f98`](https://github.com/theonlyway/recycler/commit/e880f98fdbcff4888c7d04fd16da1bea513a8330) - enhance Docker build process with SBOM generation and attestations *(commit by [@theonlyway](https://github.com/theonlyway))*

### :recycle: Refactors
- [`b2cfeae`](https://github.com/theonlyway/recycler/commit/b2cfeaec620ed5fa3fac2cdc783b498837972aec) - streamline Docker build process by replacing build-push-action with make commands and adding image digest retrieval *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.4.9] - 2026-01-12
### :wrench: Chores
- [`abdc721`](https://github.com/theonlyway/recycler/commit/abdc72121e1f9f5774eec50740b639b5bdf9b815) - **deps**: update gcr.io/distroless/static:nonroot docker digest to cba10d7 *(PR [#45](https://github.com/theonlyway/recycler/pull/45) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.4.8] - 2026-01-08
### :bug: Bug Fixes
- [`6098a40`](https://github.com/theonlyway/recycler/commit/6098a40c7990b578acc25f18136b285113f1815b) - **deps**: update go dependencies *(PR [#44](https://github.com/theonlyway/recycler/pull/44) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.4.7] - 2026-01-05
### :bug: Bug Fixes
- [`01c0e03`](https://github.com/theonlyway/recycler/commit/01c0e03e32f90b14a3f474c0ca34435ec6d87448) - add gomodTidy to postUpdateOptions in Renovate configuration [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`e17f602`](https://github.com/theonlyway/recycler/commit/e17f6023d653257fe0857fd73916e31f940d8810) - enable dependency dashboard in Renovate configuration [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`2259b27`](https://github.com/theonlyway/recycler/commit/2259b27535fa8189b07069f9a5f8ea1455c62ff8) - **deps**: pin dependencies *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.4.6] - 2026-01-05
### :bug: Bug Fixes
- [`4dedab6`](https://github.com/theonlyway/recycler/commit/4dedab65fe3f21f3c3b1e1de11d0711d92814eda) - **deps**: update kubernetes packages to v0.35.0 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`1881eb7`](https://github.com/theonlyway/recycler/commit/1881eb731d4be2d1f35b8798064429fa64dc0afe) - add rule to ignore digest updates for recycler image [skip ci[ *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`e40300c`](https://github.com/theonlyway/recycler/commit/e40300cc5af8f84931ee57a5aafe85cc1cb3357a) - update rule to ignore all updates for recycler image [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*

### :recycle: Refactors
- [`9719c79`](https://github.com/theonlyway/recycler/commit/9719c79d80df356510245840dbb523531f112d22) - update Renovate configuration to enhance security and dependency management [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`19c987d`](https://github.com/theonlyway/recycler/commit/19c987d2599d56b3c57e676cc741080ab3a22c37) - **deps**: update actions/upload-artifact action to v6 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`1486e6e`](https://github.com/theonlyway/recycler/commit/1486e6e8bc2137fc0cd11df298c6ed1933af70b5) - **deps**: update dependency go to v1.25.5 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.4.5] - 2025-12-10
### :bug: Bug Fixes
- [`c0757f3`](https://github.com/theonlyway/recycler/commit/c0757f3914d1482c1a1d963d92c9fd8f878b299d) - **deps**: update kubernetes packages to v0.34.3 *(PR [#34](https://github.com/theonlyway/recycler/pull/34) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.4.4] - 2025-12-09
### :bug: Bug Fixes
- [`7e2b1a8`](https://github.com/theonlyway/recycler/commit/7e2b1a8f081fc6b596ba231b5e6f1a7e41911671) - **deps**: update module github.com/onsi/gomega to v1.38.3 *(PR [#33](https://github.com/theonlyway/recycler/pull/33) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.4.3] - 2025-12-08
### :bug: Bug Fixes
- [`c8c4a51`](https://github.com/theonlyway/recycler/commit/c8c4a518be324bda3c51850d2d479dac7b6d6056) - **deps**: update module github.com/onsi/ginkgo/v2 to v2.27.3 *(PR [#32](https://github.com/theonlyway/recycler/pull/32) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.4.2] - 2025-12-05
### :bug: Bug Fixes
- [`9aa13ea`](https://github.com/theonlyway/recycler/commit/9aa13eaa0d28cabc05d5d83622f07a86cf27af8a) - using image tag value from semver step *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`bb6bc3a`](https://github.com/theonlyway/recycler/commit/bb6bc3aed8a577eb62dcbf358e92d4fcf6f27ae4) - added step to cleanup docker images if helm chart tests fail *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.4.1] - 2025-12-04
### :bug: Bug Fixes
- [`a1c83e4`](https://github.com/theonlyway/recycler/commit/a1c83e4ebadf52ef504be27af61a5180174b3330) - added test for helm charts *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`5ea5fdd`](https://github.com/theonlyway/recycler/commit/5ea5fdd61dd081a3c69790b8537720c63f542db0) - skip crd *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`8f343ee`](https://github.com/theonlyway/recycler/commit/8f343ee2f78c9e7be4154fb34c86efb6b9c99811) - added missing step to setup kind *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`77a1c10`](https://github.com/theonlyway/recycler/commit/77a1c107db4dac83ce40a7b9d80ba13904b04fdf) - fixed missing arg *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`e93872f`](https://github.com/theonlyway/recycler/commit/e93872f148ff6c82d2475e8272e0ed36f4564b8f) - fixed job order *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.4.0] - 2025-12-04
### :sparkles: New Features
- [`ee83051`](https://github.com/theonlyway/recycler/commit/ee83051c54321824b3f500c96558895442167fbb) - added debug condition *(commit by [@theonlyway](https://github.com/theonlyway))*

### :bug: Bug Fixes
- [`88c2787`](https://github.com/theonlyway/recycler/commit/88c2787854e0792e033cc876cb04fac40b455f65) - added badge to readme and allowed for manual trigger of schema workflow [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.3.1] - 2025-12-02
### :bug: Bug Fixes
- [`9db7bb1`](https://github.com/theonlyway/recycler/commit/9db7bb12707f95ca48951f6fdc7c1e73d3f4a7f7) - added steps to regenerate the schema for yaml-language-server *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`b60b9ee`](https://github.com/theonlyway/recycler/commit/b60b9ee792a4ba238e88b5d71e492b142e30992b) - **deps**: update actions/checkout action to v6 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`193d00d`](https://github.com/theonlyway/recycler/commit/193d00dfd17ced976a6b8ae506d38501147d995b) - **deps**: update actions/setup-go action to v6 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.3.0] - 2025-12-02
### :sparkles: New Features
- [`1285bc5`](https://github.com/theonlyway/recycler/commit/1285bc5ff84ed6a4655075c59ee27e1edf9fd106) - working on tests *(commit by [@theonlyway](https://github.com/theonlyway))*

### :bug: Bug Fixes
- [`7a857b9`](https://github.com/theonlyway/recycler/commit/7a857b91913644f82596b872ca41e7c5a89e2ba3) - fixed linter errors *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`ea0a953`](https://github.com/theonlyway/recycler/commit/ea0a953a9d8bd5deb604b6777faad1825d798001) - added logging *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`136432b`](https://github.com/theonlyway/recycler/commit/136432b2975b963834b0973a305df6132e4bac5e) - testing tests *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`cd0b147`](https://github.com/theonlyway/recycler/commit/cd0b1473ea156bfc81035fe772781ac7cff6ace1) - fixed tests and a typo *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`9b6810e`](https://github.com/theonlyway/recycler/commit/9b6810ede42228197907c93b0999c7ad46ff3c63) - added config *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`5cf4e94`](https://github.com/theonlyway/recycler/commit/5cf4e94a059330c9b0d8c3cea528e43a600cca75) - testing test changes *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`eefe57b`](https://github.com/theonlyway/recycler/commit/eefe57b60201af769f844f525e0d8a9fcc832f3b) - testing coverage *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`0b872fe`](https://github.com/theonlyway/recycler/commit/0b872fe301517d641df0fb037e8bbbde7b12f052) - don't validate crd *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`c758e79`](https://github.com/theonlyway/recycler/commit/c758e79bb8760b386f738a565059f84801f66c3c) - removed coverage html *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`fa02b7f`](https://github.com/theonlyway/recycler/commit/fa02b7fee5258caa87712becc7febb2c4343a09b) - go fmt [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`8902726`](https://github.com/theonlyway/recycler/commit/8902726fb4381896b8accafb0fc61e5688be2068) - go fmt [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.2.0] - 2025-12-02
### :sparkles: New Features
- [`ad6f9d2`](https://github.com/theonlyway/recycler/commit/ad6f9d2dcafc6d8b0146f982087d9d5a46268169) - **tests**: e2e tests actually testing pod usage *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`04d2670`](https://github.com/theonlyway/recycler/commit/04d2670fa5c67c478fa5ed15eac4c56b07aa816f) - **test**: updated test cases *(commit by [@theonlyway](https://github.com/theonlyway))*

### :bug: Bug Fixes
- [`cc25094`](https://github.com/theonlyway/recycler/commit/cc2509422b69974eb62a177cbda64402f7e4d02f) - **tests**: enabled logs *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`fe7418d`](https://github.com/theonlyway/recycler/commit/fe7418df4c62e4f1f48cf39021058ff8306b2b28) - added some verbose logging around timeout *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`6637e30`](https://github.com/theonlyway/recycler/commit/6637e300ae970fac305e683896440eba3ac6096e) - fixed linting error *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`33ab1c8`](https://github.com/theonlyway/recycler/commit/33ab1c8774dcbdd4c16e0eb82fae685e6e2a97f7) - added metrics server since kind doesn't have it by default *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`f3c834a`](https://github.com/theonlyway/recycler/commit/f3c834af9738dbf9fe1c66691f62110bd078d4d5) - additional debug *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`68af2c0`](https://github.com/theonlyway/recycler/commit/68af2c0f688e0f24573274f987d4bcfb932efa38) - changed where we fetch the CR values from *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`b47399a`](https://github.com/theonlyway/recycler/commit/b47399a3676feb90aa3accf24e8fd63df546cec2) - stupid linter max characters on a line *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`f3c13e0`](https://github.com/theonlyway/recycler/commit/f3c13e0076e8f94ee0a71f6fb36898237de3966e) - updated collection time and some kustomize overlays *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`56897ae`](https://github.com/theonlyway/recycler/commit/56897ae94310513aebd37dc5b1bd723db0cf5983) - **deps**: update actions/checkout action to v6 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`1ba8cfb`](https://github.com/theonlyway/recycler/commit/1ba8cfb93a4c7f949afac798afdaf9b416e413a2) - go fmt [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.12] - 2025-11-13
### :bug: Bug Fixes
- [`79a178e`](https://github.com/theonlyway/recycler/commit/79a178edb008c99c83b243a29a2e8d69c1561570) - **deps**: update kubernetes packages to v0.34.2 *(PR [#24](https://github.com/theonlyway/recycler/pull/24) by [@renovate[bot]](https://github.com/apps/renovate))*

### :wrench: Chores
- [`741bf5f`](https://github.com/theonlyway/recycler/commit/741bf5fc9b4901501fcd794a15b01cbb84d723b3) - **deps**: update golangci/golangci-lint-action action to v9 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.1.11] - 2025-11-03
### :bug: Bug Fixes
- [`0d0a612`](https://github.com/theonlyway/recycler/commit/0d0a61281e6bade575eff49c12af388bac6bba4f) - **deps**: update module sigs.k8s.io/controller-runtime to v0.22.4 *(PR [#22](https://github.com/theonlyway/recycler/pull/22) by [@renovate[bot]](https://github.com/apps/renovate))*

### :wrench: Chores
- [`3a158c7`](https://github.com/theonlyway/recycler/commit/3a158c7437068c39ffec42341f0c2d4c6e654fe2) - **deps**: update actions/upload-artifact action to v5 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.1.10] - 2025-10-28
### :bug: Bug Fixes
- [`a853328`](https://github.com/theonlyway/recycler/commit/a853328569885ee9de8aa5d65dbff4e31c85d0be) - **deps**: update module github.com/onsi/ginkgo/v2 to v2.27.2 *(PR [#21](https://github.com/theonlyway/recycler/pull/21) by [@renovate[bot]](https://github.com/apps/renovate))*

### :wrench: Chores
- [`6d752e8`](https://github.com/theonlyway/recycler/commit/6d752e8460bdf0cd91518f5ee391c1e036037491) - Change automergeType from 'branch' to 'pr' [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.9] - 2025-10-22
### :bug: Bug Fixes
- [`319ac85`](https://github.com/theonlyway/recycler/commit/319ac852874e0c981423973846388ed75e93caff) - **deps**: update module github.com/onsi/ginkgo/v2 to v2.27.1 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.1.8] - 2025-10-21
### :wrench: Chores
- [`0b1ef5f`](https://github.com/theonlyway/recycler/commit/0b1ef5fd8d195285f5b38b1e2e210d00a8f2335b) - Change ignoreTests from true to false *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.7] - 2025-10-20
### :wrench: Chores
- [`79be8e1`](https://github.com/theonlyway/recycler/commit/79be8e1da2ee00b0edaf550d9a6fb26c5059a2fa) - updated default values for helm chart [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`3111ab8`](https://github.com/theonlyway/recycler/commit/3111ab8e62a4734c994f61702b8d3c21a556a722) - updated kustomization to make renovate happy *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.6] - 2025-10-20
### :wrench: Chores
- [`bcd469f`](https://github.com/theonlyway/recycler/commit/bcd469f4bd0b0b299a7e5bcd54ecd648c275a3d7) - **deps**: testing pipeline changes *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`e70c152`](https://github.com/theonlyway/recycler/commit/e70c152fa8ced7ed26e46100d873e6aa35b6b36d) - **deps**: removed executing on pr *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`7e7a9bf`](https://github.com/theonlyway/recycler/commit/7e7a9bf39ed59805a73264dd3a81dc2c25291536) - **deps**: updated pipeline formatting *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`108f5e4`](https://github.com/theonlyway/recycler/commit/108f5e4b402b5d94cffd5dda7f342c1a4fab6e52) - **deps**: allowing pipeline to be manually executed *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`44d56fb`](https://github.com/theonlyway/recycler/commit/44d56fbf32ec2b455edbb388f44f49e03e16aec2) - **deps**: renamed workflow *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`fcbb65c`](https://github.com/theonlyway/recycler/commit/fcbb65cffbb291f2bc10475ee9af2959b64075c3) - **deps**: working on build workflow *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`7c00063`](https://github.com/theonlyway/recycler/commit/7c000636ad6c6874b51fc6cfd92ea4f55e01d9b7) - **deps**: to slow for testing *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`222b62d`](https://github.com/theonlyway/recycler/commit/222b62d0e1c8a8d9ebc821114cd0c96e4c3d2d5c) - **deps**: working on release *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`b7102e5`](https://github.com/theonlyway/recycler/commit/b7102e53be65245c5837caa5094245edcd554278) - updated readme [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`5bfc807`](https://github.com/theonlyway/recycler/commit/5bfc807d1e9245f2f8fb41921efb8625df949ba8) - Update workflow dependencies for builds and releases *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.5] - 2025-10-20
### :bug: Bug Fixes
- [`8c0c472`](https://github.com/theonlyway/recycler/commit/8c0c472f996ab5496a85b6999b4ff9b79092f5a0) - always the little things *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.4] - 2025-10-20
### :bug: Bug Fixes
- [`8ea5569`](https://github.com/theonlyway/recycler/commit/8ea55690c9023be0c6fcffac035eaefa42b5274b) - added namespace.yaml *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.3] - 2025-10-20
### :bug: Bug Fixes
- [`aa1bd4c`](https://github.com/theonlyway/recycler/commit/aa1bd4ccb403e4fef8adf904371b5f88ec6b4d89) - fixed helm charts *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`3e16a24`](https://github.com/theonlyway/recycler/commit/3e16a24878f2153f43b1f8182c3832f727612823) - **deps**: update actions/checkout action to v5 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`fc9d08d`](https://github.com/theonlyway/recycler/commit/fc9d08d9a1dfb5e2df2e218386d7ee6bfa1da177) - **deps**: update actions/setup-go action to v6 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.1.2] - 2025-10-20
### :bug: Bug Fixes
- [`20d82e1`](https://github.com/theonlyway/recycler/commit/20d82e104a54cab7586413b18153ff723809aa9a) - fixed indenting *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.1] - 2025-10-20
### :bug: Bug Fixes
- [`b1fc00a`](https://github.com/theonlyway/recycler/commit/b1fc00a737f6229ae9759b774a50fa9d0b76e50a) - updated namespace in helm charts *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.1.0] - 2025-10-20
### :sparkles: New Features
- [`a3c871f`](https://github.com/theonlyway/recycler/commit/a3c871fbcf253d06a0b9b9dfdcfd96c7900046ce) - working on tests and updated recycler namespace *(commit by [@theonlyway](https://github.com/theonlyway))*

### :bug: Bug Fixes
- [`bdf6f7c`](https://github.com/theonlyway/recycler/commit/bdf6f7c8f59e1331fc005ca1b0b4d8ec05193aed) - fixed tests for pipeline *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`56417cc`](https://github.com/theonlyway/recycler/commit/56417cca4956839e2d4d9ebb92b8fc907148eb6a) - fixed docker-build *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`96b3d9b`](https://github.com/theonlyway/recycler/commit/96b3d9b49182e01575379aa77b0563af4018615c) - removed platform *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`4a37081`](https://github.com/theonlyway/recycler/commit/4a370812bbe1ca6523005eebeb76bf5151aa7383) - updated go version to match docker image *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`e021800`](https://github.com/theonlyway/recycler/commit/e021800a6fb8e410cf4eb39b880375e7b759c3ca) - updated go *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`5621c4a`](https://github.com/theonlyway/recycler/commit/5621c4a8629f8fecff9e6ac9b3130c04ee141074) - that shouldn't be missing apparently? *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`0a1572e`](https://github.com/theonlyway/recycler/commit/0a1572e27e9b7e51878e13399a95ee00a1af903d) - updated go packages *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`536cac9`](https://github.com/theonlyway/recycler/commit/536cac9b639aee7a21a1f1a327189948e9c6e043) - updated buildx command *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`a83e3ac`](https://github.com/theonlyway/recycler/commit/a83e3acb8755d7cf5e9038361736b1bcf53f5be7) - updated controller gen *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`a3450bd`](https://github.com/theonlyway/recycler/commit/a3450bde8caa3c7f2b4cc78be79c4738dc7b2796) - testing in pipeline *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`c7c5077`](https://github.com/theonlyway/recycler/commit/c7c5077344a91685996f84fadc247e93fbe80563) - updated tests *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`f92cf1e`](https://github.com/theonlyway/recycler/commit/f92cf1e29bc36fbb0eae6a5675df4bae74d42959) - added coverage artifact *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`1e933c4`](https://github.com/theonlyway/recycler/commit/1e933c4f28b2defc48cafaf2b571847f55a01a1a) - test version upgrade *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`ac47855`](https://github.com/theonlyway/recycler/commit/ac47855e3cb7142e9c59d13a08b6c7b244b00b2e) - resolved linting errors *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`ed985a1`](https://github.com/theonlyway/recycler/commit/ed985a1c58c00e5bb824fc4f6690547f04a4a39a) - import in wrong spot *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`4e931fb`](https://github.com/theonlyway/recycler/commit/4e931fb31b89ef1ec38505447ecefd786292e23f) - fix formatting *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`85762c6`](https://github.com/theonlyway/recycler/commit/85762c6ac3a2165a526be05bc9903e89ef2a0e0e) - added extra workflows *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`a10faac`](https://github.com/theonlyway/recycler/commit/a10faac6608f39296807b7e0f8cc89f3f6511cbc) - added other badges [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`cff7cea`](https://github.com/theonlyway/recycler/commit/cff7cea9f609fc624f68536905e5441c7c73e0e4) - **config**: migrate config renovate.json *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.0.8] - 2025-10-17
### :bug: Bug Fixes
- [`8138a54`](https://github.com/theonlyway/recycler/commit/8138a54910f36fec312d30f177f14bdb22078da9) - **deps**: update module sigs.k8s.io/controller-runtime to v0.22.3 *(PR [#9](https://github.com/theonlyway/recycler/pull/9) by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.0.7] - 2025-10-17
### :bug: Bug Fixes
- [`1ea34f6`](https://github.com/theonlyway/recycler/commit/1ea34f63e704f0bcd6ef7b83fb72e1648bd3caf4) - **deps**: update kubernetes packages to v0.34.1 *(PR [#6](https://github.com/theonlyway/recycler/pull/6) by [@renovate[bot]](https://github.com/apps/renovate))*

### :wrench: Chores
- [`a7526cd`](https://github.com/theonlyway/recycler/commit/a7526cd609f148fb3caaad6775d125531b5e9fc4) - **deps**: update actions/checkout action to v5 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`9fcda7b`](https://github.com/theonlyway/recycler/commit/9fcda7b493dc57c637a4d75f1a4b38c69ed4084c) - **deps**: update actions/setup-go action to v6 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`3c44255`](https://github.com/theonlyway/recycler/commit/3c4425527a65658be06d65cfaba93145f84033a9) - **deps**: update actions/upload-pages-artifact action to v4 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`10a7843`](https://github.com/theonlyway/recycler/commit/10a78439b05e667d1ab7a4d9a5edd338faba63d4) - **deps**: update stefanzweifel/git-auto-commit-action action to v7 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*
- [`8b67dbf`](https://github.com/theonlyway/recycler/commit/8b67dbf5a415782d8ce609e7bee4287dcf2c1ee7) - **deps**: update clementtsang/delete-tag-and-release action to v0.4.0 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.0.6] - 2025-10-17
### :bug: Bug Fixes
- [`82a2666`](https://github.com/theonlyway/recycler/commit/82a2666fbbd43da48331b8188e6f5abc1dd71103) - **deps**: update module github.com/go-logr/logr to v1.4.3 *(PR [#2](https://github.com/theonlyway/recycler/pull/2) by [@renovate[bot]](https://github.com/apps/renovate))*
- [`b6060c6`](https://github.com/theonlyway/recycler/commit/b6060c6149b8ee74f903cae59df6750a86218f71) - **deps**: update module github.com/onsi/ginkgo/v2 to v2.26.0 *(PR [#7](https://github.com/theonlyway/recycler/pull/7) by [@renovate[bot]](https://github.com/apps/renovate))*

### :wrench: Chores
- [`f1262eb`](https://github.com/theonlyway/recycler/commit/f1262eb17fb6e76f208f50b9f49ab6634e9cc183) - switched automerge type *(commit by [@theonlyway](https://github.com/theonlyway))*
- [`737be54`](https://github.com/theonlyway/recycler/commit/737be54a3a599edc438e4393610f56253943f422) - added no tests [skip ci] *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.0.5] - 2025-10-17
### :bug: Bug Fixes
- [`5a1edf5`](https://github.com/theonlyway/recycler/commit/5a1edf5ab5ff4cd70cb0cd7115995c19920c8f0e) - updated push logic *(commit by [@theonlyway](https://github.com/theonlyway))*

### :wrench: Chores
- [`df89eb8`](https://github.com/theonlyway/recycler/commit/df89eb8f5da56b5976f8dac867f9b585acee367f) - **deps**: update golang docker tag to v1.25 *(commit by [@renovate[bot]](https://github.com/apps/renovate))*


## [1.0.4] - 2025-10-17
### :wrench: Chores
- [`cc58cad`](https://github.com/theonlyway/recycler/commit/cc58cad4a86cdf2640f592e25a8d405dc3be37dc) - Update renovate.json with new settings *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.0.3] - 2025-10-17
### :bug: Bug Fixes
- [`4f79cdf`](https://github.com/theonlyway/recycler/commit/4f79cdf9f273b0ea1abc75dccc129de839a51441) - fixed default container name *(commit by [@theonlyway](https://github.com/theonlyway))*


## [1.0.1] - 2025-05-22
### :bug: Bug Fixes
- [`2348672`](https://github.com/theonlyway/recycler/commit/2348672a791b5f7040deab73a47749b8afbc9f54) - building release *(commit by [@theonlyway](https://github.com/theonlyway))*


[1.0.1]: https://github.com/theonlyway/recycler/compare/v1.0.0...1.0.1
[1.0.3]: https://github.com/theonlyway/recycler/compare/1.0.2...1.0.3
[1.0.4]: https://github.com/theonlyway/recycler/compare/1.0.3...1.0.4
[1.0.5]: https://github.com/theonlyway/recycler/compare/1.0.4...1.0.5
[1.0.6]: https://github.com/theonlyway/recycler/compare/1.0.5...1.0.6
[1.0.7]: https://github.com/theonlyway/recycler/compare/1.0.6...1.0.7
[1.0.8]: https://github.com/theonlyway/recycler/compare/1.0.7...1.0.8
[1.1.0]: https://github.com/theonlyway/recycler/compare/1.0.8...1.1.0
[1.1.1]: https://github.com/theonlyway/recycler/compare/1.1.0...1.1.1
[1.1.2]: https://github.com/theonlyway/recycler/compare/1.1.1...1.1.2
[1.1.3]: https://github.com/theonlyway/recycler/compare/1.1.2...1.1.3
[1.1.4]: https://github.com/theonlyway/recycler/compare/1.1.3...1.1.4
[1.1.5]: https://github.com/theonlyway/recycler/compare/1.1.4...1.1.5
[1.1.6]: https://github.com/theonlyway/recycler/compare/1.1.5...1.1.6
[1.1.7]: https://github.com/theonlyway/recycler/compare/1.1.6...1.1.7
[1.1.8]: https://github.com/theonlyway/recycler/compare/1.1.7...1.1.8
[1.1.9]: https://github.com/theonlyway/recycler/compare/1.1.8...1.1.9
[1.1.10]: https://github.com/theonlyway/recycler/compare/1.1.9...1.1.10
[1.1.11]: https://github.com/theonlyway/recycler/compare/1.1.10...1.1.11
[1.1.12]: https://github.com/theonlyway/recycler/compare/1.1.11...1.1.12
[1.2.0]: https://github.com/theonlyway/recycler/compare/1.1.12...1.2.0
[1.3.0]: https://github.com/theonlyway/recycler/compare/1.2.0...1.3.0
[1.3.1]: https://github.com/theonlyway/recycler/compare/1.3.0...1.3.1
[1.4.0]: https://github.com/theonlyway/recycler/compare/1.3.1...1.4.0
[1.4.1]: https://github.com/theonlyway/recycler/compare/1.4.0...1.4.1
[1.4.2]: https://github.com/theonlyway/recycler/compare/1.4.1...1.4.2
[1.4.3]: https://github.com/theonlyway/recycler/compare/1.4.2...1.4.3
[1.4.4]: https://github.com/theonlyway/recycler/compare/1.4.3...1.4.4
[1.4.5]: https://github.com/theonlyway/recycler/compare/1.4.4...1.4.5
[1.4.6]: https://github.com/theonlyway/recycler/compare/1.4.5...1.4.6
[1.4.7]: https://github.com/theonlyway/recycler/compare/1.4.6...1.4.7
[1.4.8]: https://github.com/theonlyway/recycler/compare/1.4.7...1.4.8
[1.4.9]: https://github.com/theonlyway/recycler/compare/1.4.8...1.4.9
[1.5.0]: https://github.com/theonlyway/recycler/compare/1.4.9...1.5.0
[1.5.1]: https://github.com/theonlyway/recycler/compare/1.5.0...1.5.1
[1.6.0]: https://github.com/theonlyway/recycler/compare/1.5.1...1.6.0
[1.6.1]: https://github.com/theonlyway/recycler/compare/1.6.0...1.6.1
[1.6.2]: https://github.com/theonlyway/recycler/compare/1.6.1...1.6.2
[1.6.3]: https://github.com/theonlyway/recycler/compare/1.6.2...1.6.3
[1.6.4]: https://github.com/theonlyway/recycler/compare/1.6.3...1.6.4
[1.6.5]: https://github.com/theonlyway/recycler/compare/1.6.4...1.6.5
[1.6.6]: https://github.com/theonlyway/recycler/compare/1.6.5...1.6.6
[1.7.1]: https://github.com/theonlyway/recycler/compare/1.7.0...1.7.1
[1.8.2]: https://github.com/theonlyway/recycler/compare/1.8.1...1.8.2
[1.9.0]: https://github.com/theonlyway/recycler/compare/1.8.2...1.9.0
[1.9.1]: https://github.com/theonlyway/recycler/compare/1.9.0...1.9.1
[1.9.2]: https://github.com/theonlyway/recycler/compare/1.9.1...1.9.2
