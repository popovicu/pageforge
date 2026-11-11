"""site_build and site_zip rules."""

load("@rules_pkg//pkg:providers.bzl", "PackageFilesInfo")
load("@rules_pkg//pkg:tar.bzl", "pkg_tar")
load(":providers.bzl", "SiteOutputInfo")

def site_build(name, deps = [], static = [], collections = [], visibility = None, **kwargs):
    """Collect all SiteOutputInfo entries and assemble into a tar archive.

    Args:
        name: Target name. Produces <name>.tar.
        deps: Targets providing SiteOutputInfo (pages).
        static: Targets providing SiteOutputInfo (static files).
        collections: Targets producing .tar archives (collections, tags).
        visibility: Bazel visibility.
        **kwargs: Passed through to pkg_tar.
    """
    collect_name = name + "_collect"

    _site_collect(
        name = collect_name,
        deps = deps + static,
        collections = collections,
    )

    pkg_tar(
        name = name,
        srcs = [":" + collect_name],
        deps = collections,
        allow_duplicates_with_different_content = False,
        visibility = visibility,
        **kwargs
    )

def _label_str(label):
    s = str(label)
    return s[2:] if s.startswith("@@//") else s

def _site_collect_impl(ctx):
    dest_src_map = {}
    owners = {}
    for dep in ctx.attr.deps:
        for entry in dep[SiteOutputInfo].entries:
            url = entry.url_path.lstrip("/")
            prev = dest_src_map.get(url)
            if prev != None and prev != entry.file:
                fail("site_build: URL '/%s' is produced by both %s and %s" % (url, prev.short_path, entry.file.short_path))
            dest_src_map[url] = entry.file
            owners[url] = _label_str(dep.label)

    # pkg_tar merges collection tars by keeping the first copy of a path and
    # only printing a warning, so cross-channel collisions are checked here,
    # in a validation action that fails the build instead.
    tars = []
    for target in ctx.attr.collections:
        for f in target[DefaultInfo].files.to_list():
            if f.basename.endswith(".tar"):
                tars.append(struct(label = _label_str(target.label), path = f.path, file = f))
    validation = []
    if tars:
        manifest = ctx.actions.declare_file(ctx.label.name + ".check.json")
        ctx.actions.write(manifest, json.encode({
            "files": owners,
            "tars": [{"label": t.label, "path": t.path} for t in tars],
        }))
        stamp = ctx.actions.declare_file(ctx.label.name + ".check")
        ctx.actions.run(
            executable = ctx.executable._sitecheck,
            arguments = ["--manifest", manifest.path, "--output", stamp.path],
            inputs = [manifest] + [t.file for t in tars],
            outputs = [stamp],
            mnemonic = "SiteCheck",
            progress_message = "Checking %{label} for URL collisions",
        )
        validation.append(stamp)

    return [
        PackageFilesInfo(
            dest_src_map = dest_src_map,
            attributes = {"mode": "0644"},
        ),
        DefaultInfo(files = depset(dest_src_map.values())),
        OutputGroupInfo(_validation = depset(validation)),
    ]

_site_collect = rule(
    implementation = _site_collect_impl,
    attrs = {
        "deps": attr.label_list(providers = [SiteOutputInfo]),
        "collections": attr.label_list(),
        "_sitecheck": attr.label(
            default = "//renderers/sitecheck",
            executable = True,
            cfg = "exec",
        ),
    },
)

def _site_zip_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".zip")
    src = ctx.file.tar
    ctx.actions.run_shell(
        outputs = [out],
        inputs = [src],
        arguments = [src.path, out.path],
        command = """
set -eu
IN=$1
OUT=$2
case "$OUT" in /*) ;; *) OUT="$PWD/$OUT" ;; esac
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
tar -xf "$IN" -C "$TMP"
(cd "$TMP" && zip -qrX "$OUT" .)
""",
        mnemonic = "SiteZip",
        progress_message = "Repacking %{label} as zip",
        use_default_shell_env = True,
    )
    return [DefaultInfo(files = depset([out]))]

site_zip = rule(
    implementation = _site_zip_impl,
    doc = "Repack a site_build .tar archive as a .zip. The `tar` attribute names " +
          "the source archive explicitly: this rule does not build a site from " +
          "scratch, it just rewraps an existing tar. Use for upload targets that " +
          "require zip (e.g. Cloudflare Pages).",
    attrs = {
        "tar": attr.label(
            allow_single_file = [".tar"],
            mandatory = True,
            doc = "site_build target whose .tar output should be repacked as zip.",
        ),
    },
)
