---
title: "What a UNIX pipe is, exactly"
date: 2026-04-01
description: "A pipe is a kernel buffer with two file descriptors. Everything else follows."
tags: [unix, kernel]
author: "Jules"
reading_minutes: 5
---

A pipe is two file descriptors and a kernel-managed circular buffer between
them. Write to one end, read from the other. The kernel does the queueing,
the blocking, and the EOF signaling.

Everything you can say about pipes follows from that one sentence.

## The system call

```c
int fds[2];
if (pipe(fds) < 0) {
    perror("pipe");
    return 1;
}
// fds[0] is the read end. fds[1] is the write end.
```

The kernel allocates a buffer (64 KiB on Linux by default) and gives you two
descriptors that share it. The buffer survives until both descriptors are
closed.

## Why two ends, not one

The asymmetry exists so the kernel can detect EOF. When every descriptor
referring to the *write* end is closed, a `read` on the *read* end returns
zero — that is what every shell program reads as "the input ended."

If a single bidirectional fd were used, there would be nothing to close, and
no way to signal "no more bytes coming."

## Backpressure

The buffer has a fixed size. When it fills, the next `write` blocks until a
reader makes room. When it empties, the next `read` blocks until a writer
adds bytes.

This is the entire mechanism behind shell pipelines like `grep ... | head -n
1` exiting cleanly: when `head` closes its read end, the next `write` from
`grep` gets `SIGPIPE` and the process exits.

## What a shell does with this

```bash
$ ls | wc -l
```

Behind the scenes the shell does roughly:

```c
int fds[2];
pipe(fds);

if (fork() == 0) {
    dup2(fds[1], STDOUT_FILENO);
    close(fds[0]);
    close(fds[1]);
    execlp("ls", "ls", NULL);
}

if (fork() == 0) {
    dup2(fds[0], STDIN_FILENO);
    close(fds[0]);
    close(fds[1]);
    execlp("wc", "wc", "-l", NULL);
}

close(fds[0]);
close(fds[1]);
wait(NULL);
wait(NULL);
```

That is the entire trick. `dup2` rewires the descriptors before `exec`, the
parent closes its copies (otherwise the read end never sees EOF), and then
both children just do their normal stdin/stdout thing.

A shell is, to a first approximation, an extremely careful caller of `pipe`
and `dup2`.
