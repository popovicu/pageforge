"""site_serve rule.

Creates a `bazel run` target that serves a built site locally over HTTP.
"""

def _site_serve_impl(ctx):
    tar_file = ctx.attr.site[DefaultInfo].files.to_list()[0]

    # Write a launcher script that invokes the devserver with the tar path.
    launcher = ctx.actions.declare_file(ctx.label.name + "_launcher.sh")
    ctx.actions.write(
        output = launcher,
        content = """#!/bin/bash
set -e
SITE_DIR=$(mktemp -d)
trap 'rm -rf "$SITE_DIR"' EXIT
tar xf {tar} -C "$SITE_DIR"
# Not exec: the shell must outlive the server so the EXIT trap removes
# SITE_DIR. Stop signals sent to this script are forwarded to the server.
{devserver} --dir "$SITE_DIR" --port {port} "$@" &
SERVER=$!
trap 'kill "$SERVER" 2>/dev/null' INT TERM HUP
wait "$SERVER"
""".format(
            devserver = ctx.executable._devserver.short_path,
            tar = tar_file.short_path,
            port = ctx.attr.port,
        ),
        is_executable = True,
    )

    runfiles = ctx.runfiles(files = [tar_file])
    runfiles = runfiles.merge(ctx.attr._devserver[DefaultInfo].default_runfiles)

    return [
        DefaultInfo(
            executable = launcher,
            runfiles = runfiles,
        ),
    ]

site_serve = rule(
    implementation = _site_serve_impl,
    doc = "Serve a built site locally via `bazel run`. Starts an HTTP file server.",
    executable = True,
    attrs = {
        "site": attr.label(
            mandatory = True,
            doc = "A site_build target producing a .tar archive.",
        ),
        "port": attr.int(
            default = 8080,
            doc = "Port to listen on.",
        ),
        "_devserver": attr.label(
            default = "//renderers/devserver",
            executable = True,
            cfg = "target",
        ),
    },
)
