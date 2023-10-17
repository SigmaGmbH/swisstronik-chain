workspace(name = "swisstronik_chain")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "com_google_protobuf",
    sha256 = "5980276108f948e1ada091475549a8c75dc83c193129aab0e986ceaac3e97131",
    strip_prefix = "protobuf-24.0",
    urls = ["https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protobuf-24.0.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "91585017debb61982f7054c9688857a2ad1fd823fc3f9cb05048b0025c47d023",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.42.0/rules_go-v0.42.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.42.0/rules_go-v0.42.0.zip",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.21.1")

http_archive(
    name = "bazel_gazelle",
    sha256 = "d3fa66a39028e97d76f9e2db8f1b0c11c099e8e01bf363a923074784e451f809",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.33.0/bazel-gazelle-v0.33.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.33.0/bazel-gazelle-v0.33.0.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
gazelle_dependencies()

load("//:go_repositories.bzl", "go_repositories")

# gazelle:repository_macro go_repositories.bzl%go_repositories
go_repositories()

go_repository(
    name = "com_github_cosmos_cosmos_sdk",
    importpath = "github.com/cosmos/cosmos-sdk",
    replace = "github.com/cosmos/cosmos-sdk",
    sum = "h1:LhL6WDBadczqBuCW0t5BHUzGQR3vbujdOYOfU0ORt+o=",
    version = "v0.46.13",
    build_directives = [
        "gazelle:proto_strip_import_prefix /proto",
        "gazelle:resolve proto gogoproto/gogo.proto @com_github_regen_network_protobuf//gogoproto:gogoproto_proto",
        "gazelle:resolve proto go gogoproto/gogo.proto @com_github_regen_network_protobuf//gogoproto:go_default_library",
        "gazelle:resolve proto cosmos_proto/cosmos.proto @com_github_cosmos_cosmos_proto//proto/cosmos_proto:cosmos_proto_proto",
        "gazelle:resolve proto go cosmos_proto/cosmos.proto @com_github_cosmos_cosmos_proto//proto/cosmos_proto:go_default_library",
        "gazelle:resolve_regexp proto google/api/.*\\.proto @com_github_gogo_googleapis//google/api:api_proto",
        "gazelle:resolve_regexp proto go google/api/.*\\.proto @com_github_gogo_googleapis//google/api:go_default_library",
    ],
)
go_repository(
    name = "com_github_cosmos_cosmos_proto",
    importpath = "github.com/cosmos/cosmos-proto",
    sum = "h1:iDL5qh++NoXxG8hSy93FdYJut4XfgbShIocllGaXx/0=",
    version = "v1.0.0-beta.1",
    build_directives = [
        "gazelle:proto_strip_import_prefix /proto",
    ],
)

load("//:go_repositories_extra.bzl", "go_repositories_extra")
go_repositories_extra()
