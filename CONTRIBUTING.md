# Contributing

## License

By contributing you agree to license your contribution under the terms of the
[Apache License 2.0](LICENSE).

## Reporting Bugs

When reporting bugs please ensure you include all necessary information to reproduce
the bug. Bug reports that include a unit test that demonstrates the bug will get more
attention from maintainers.

If you are absolutely certain that what you found is a bug, you can reproduce it in a unit
test and it is obvious to you that the bug must be fixed, you can skip the issue creation
phase and submit a PR that fixes the bug. Please be aware that the PR may be rejected if
it does not meet maintainers' expectations (e.g. is not considered to be a bug by
maintainers).

## Fixing Bugs

All PRs that fix a bug must include an automated test that demonstrates the bug. The test
must fail _before_ the fix and must pass _after_ the fix.

## Pull Requests

All PRs must contain a single commit. The commit message is required. One-line commit
message are only acceptable for absolutely trivial and self-explanatory changes. For most
other changes a detailed commit description is required. See
[how to write good commit messages](https://cbea.ms/git-commit/).

If the PR is related to an already existing issue (as it should in most cases) make sure
to link to the issue using an appropriate
[keyword](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue).
Merging the PR must auto-close the relevant issue.

PRs must include automated tests that verify the added/changed functionality.

If you need to make changes to an existing PR and want the reviewers to clearly see
how the PR changed compared to its initial state push additional commits to the PR's
branch. If it is not important for the PR delta to be clearly visible then it is
acceptable to amend the original commit.

We aim for linear git commit history and use squash and merge without merge commits.

Each PR must do only one thing: one bug fix, one feature addition or one 
logical refactoring, etc. Don't pile up changes in one PR. This is to ensure individual 
commits can be reverted easily if needed without impact any other change. 

Large changes may be broken down into several PRs, however each PR must be a 
logically reasonable piece of work, must build correctly and preferably have no
effect at runtime until all PRs that comprise the change are merged. A possible approach
is to start with PRs that add functionality that is well-tested but is not callable 
externally and then make the functionality accessible in the last PR that introduces the 
public API to the added functionality.

## Automated Tests

We care about automated test suite execution speed. Please be mindful of tests that
slow down the test suite significantly.

## Low Quality PRs

Low effort, low quality PRs (often ones created using Gen AI) will not be accepted.

If you would like to contribute to the project please make an effort to read and meet
expectations of this contributing guide and general expectations of high quality software
engineering work.

## New Feature Proposals

Please start by creating a new issue that describes the feature. The issue will be 
used to discuss decide if the feature will be accepted. Do not create a PR with 
feature implementation without first discussing the feature with maintainers. There is 
a good chance the PR will be rejected and your effort will be wasted.

We expect well-thought-out proposals that describe the use cases, the motivation, any 
tradeoffs involved, alternates considered, etc. For complicated, major proposals you will 
want to attach a full-blown design document.

Any additions or changes to the format must include a prototype implementation of the
change that demonstrates how it can be implemented and what the performance impact of the
change is. The prototype may omit some elements that we will require later if the feature
is accepted (e.g. automated tests, documentation, etc).

## Performance

We care a lot about performance:

- Speed (CPU usage).
- Memory usage and memory allocations.
- Payload size in uncompressed and especially in zstd compressed form.

All changes must be evaluated from performance perspective. PRs automatically run
benchmarks before and after the change and publish a benchmark diff in github action 
log to make it easy to assess performance impact. You can and should also do 
benchmarking locally before creating the PR, see [benchmarks](./benchmarks) directory.

PRs that are otherwise fine may be rejected if they degrade performance.

## Coding Guidelines

For Go code follow [Effective Go](https://go.dev/doc/effective_go) recommendations.

### Comments

Make sure to document your code via comments where necessary. Here are some 
[recommendations](https://antirez.com/news/124) on how to write good comments.
Hint: you don't want Trivial or Backup comments.

All public API (e.g. exported Go functions) must be documented.

Code generated by `stefgen` must be nicely readable. Make sure the generator 
templates include proper comments.

## Releasing

This section contains instructions for maintainers only.

1. To release a new version vX.Y.Z run `make prepver VERSION=vX.Y.Z` to update go.mod
   files.
2. Create a PR with changes and get it merged.
3. Run `make releasever VERSION=vX.Y.Z` to create and push version tags to github.
4. Create a [new release](https://github.com/splunk/stef/releases) with the same version
   vX.Y.Z and make sure to include a changelog. You can use "Generate release notes"
   button, but make sure to tidy up and get rid of changes that are not important from
   external observer perspective. Changelog is intended to be read by users of this repo.
5. Make sure no new PRs are merged between step 1 and 5. The release and all tags MUST
   point to the same commit.
