load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_repositories_extra():
    go_repository(
        name = "com_github_regen_network_protobuf",
        importpath = "github.com/regen-network/protobuf",
        sum = "h1:OHEc+q5iIAXpqiqFKeLpu5NwTIkVXUs48vFMwzqpqY4=",
        version = "v1.3.3-alpha.regen.1",
    )