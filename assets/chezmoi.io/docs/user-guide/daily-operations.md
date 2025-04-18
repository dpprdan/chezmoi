# Daily operations

## Edit your dotfiles

Edit a dotfile with:

```sh
chezmoi edit $FILENAME
```

This will edit `$FILENAME`'s source file in your source directory. chezmoi will
not make any changes to the actual dotfile until you run `chezmoi apply`.

To automatically run `chezmoi apply` when you quit your editor, run:

```sh
chezmoi edit --apply $FILENAME
```

To automatically run `chezmoi apply` whenever you save the file in your editor, run:

```sh
chezmoi edit --watch $FILENAME
```

You don't have to use `chezmoi edit` to edit your dotfiles. For more
information, see [Do I have to use `chezmoi edit` to edit my
dotfiles?][faq-edit].

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    W->>W: chezmoi edit
    W->>H: chezmoi apply
    W->>H: chezmoi edit --apply
    W->>H: chezmoi edit --watch
```

## Pull the latest changes from your repo and apply them

You can pull the changes from your repo and apply them in a single command:

```sh
chezmoi update
```

This runs `git pull --autostash --rebase` in your source directory and then
`chezmoi apply`.

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    R->>H: chezmoi update
```

## Pull the latest changes from your repo and see what would change, without actually applying the changes

Run:

```sh
chezmoi git pull -- --autostash --rebase && chezmoi diff
```

This runs `git pull --autostash --rebase` in your source directory and `chezmoi
diff` then shows the difference between the target state computed from your
source directory and the actual state.

If you're happy with the changes, then you can run

```sh
chezmoi apply
```

to apply them.

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    R->>W: chezmoi git pull
    W-->>H: chezmoi diff
    W->>H: chezmoi apply
```

## Automatically commit and push changes to your repo

chezmoi can automatically commit and push changes to your source directory to
your repo. This feature is disabled by default. To enable it, add the following
to your config file:

```toml title="~/.config/chezmoi/chezmoi.toml"
[git]
    autoCommit = true
    autoPush = true
```

Whenever a change is made to your source directory, chezmoi will commit the
changes with an automatically-generated commit message (if `autoCommit` is true)
and push them to your repo (if `autoPush` is true). `autoPush` implies
`autoCommit`, i.e. if `autoPush` is true then chezmoi will auto-commit your
changes. If you only set `autoCommit` to true then changes will be committed but
not pushed.

By default, `autoCommit` will generate a commit message based on the files
changed. You can override this by setting the `git.commitMessageTemplate`
configuration variable. For example, to have chezmoi prompt you for a commit
message each time, use:

```toml title="~/.config/chezmoi/chezmoi.toml"
[git]
    autoCommit = true
    commitMessageTemplate = "{{ promptString \"Commit message\" }}"
```

If your commit message is longer than fits in a string then you can set
`git.commitMessageTemplateFile` to specify a path to the commit message template
relative to the source directory, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[git]
    autoCommit = true
    commitMessageTemplateFile = ".commit_message.tmpl"
```

Be careful when using `autoPush`. If your dotfiles repo is public and you
accidentally add a secret in plain text, that secret will be pushed to your
public repo.

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    W->>L: autoCommit
    W->>R: autoPush
```

## Install chezmoi and your dotfiles on a new machine with a single command

chezmoi's install script can run `chezmoi init` for you by passing extra
arguments to the newly installed chezmoi binary. If your dotfiles repo is
`github.com/$GITHUB_USERNAME/dotfiles` then installing chezmoi, running
`chezmoi init`, and running `chezmoi apply` can be done in a single line of
shell:

```sh
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply $GITHUB_USERNAME
```

If your dotfiles repo has a different name to `dotfiles`, or if you host your
dotfiles on a different service, then see the [reference manual for `chezmoi
init`][init].

For setting up transitory environments (e.g. short-lived Linux containers) you
can install chezmoi, install your dotfiles, and then remove all traces of
chezmoi, including the source directory and chezmoi's configuration directory,
with a single command:

```sh
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --one-shot $GITHUB_USERNAME
```

```mermaid
sequenceDiagram
    participant H as home directory
    participant W as working copy
    participant L as local repo
    participant R as remote repo
    R->>W: chezmoi init $GITHUB_USERNAME
    R->>H: chezmoi init --apply $GITHUB_USERNAME
    R->>H: chezmoi init --one-shot $GITHUB_USERNAME
```

[faq-edit]: /user-guide/frequently-asked-questions/usage.md#how-do-i-edit-my-dotfiles-with-chezmoi
[init]: /reference/commands/init.md
