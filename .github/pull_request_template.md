## Summary

<!-- What changed and why. Reviewers should understand the purpose from this
     section alone. Use bullets. Link the issue if there is one. -->

Resolves #

## Demo

<!-- Optional. Before/after output for `gh repox create`, `apply`, or `diff`
     (dry-run is enough). Delete this section if it would not help a reviewer. -->

## Test plan

<!-- Fill in what you actually ran. Do not check an item unless you ran that
     command and looked at the output.

     Confirm the test really ran. A green suite is not enough if the new test
     never executed (wrong package, skip, or filter). The test name should
     appear in the output.

     For new or changed coverage, also confirm failure: watch the test fail
     for the expected reason (missing behavior or the bug), then pass after
     the change. If it passed on the first run, it may not be testing what
     you think.

     If a test might be flaky (timing, concurrency, network, filesystem,
     ordering), attach a flakiness sweep below. -->

- [ ]
- [ ] `make check` is green

### Flakiness sweep

<!-- Required only if a new or changed test might be flaky. Re-run it enough
     times to be convincing, for example:

       go test -count=100 ./internal/github

     Paste the command and a pass summary (N/N). Delete this subsection
     otherwise. -->

## Reviewer notes

<!-- Optional. Risky files, tricky logic, or decisions you want a second
     opinion on. Delete if none. -->
