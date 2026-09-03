# Maintenance

## Publishing a release

Releases use a manually created tag on a commit in `main`. Pushing a tag that starts with `v` starts the `release-version` GitHub Actions workflow. The workflow runs the tests and uses GoReleaser to publish the release.

1. Update the local `main` branch and fetch all tags:

   ```fish
   git switch main
   git pull --ff-only origin main
   git fetch origin --tags
   ```

2. Find the latest release and verify all changes since that release:

   ```fish
   set previous_tag (git describe --tags --abbrev=0)
   git log --oneline "$previous_tag"..HEAD
   git diff --stat "$previous_tag"..HEAD
   ```

   Confirm that the selected commit is in `main`, that the changes are complete, and that the checks for the commit passed.

3. Select the new version according to [Semantic Versioning](https://semver.org/):

   - Increment **major** for incompatible changes.
   - Increment **minor** for backward-compatible functionality.
   - Increment **patch** for backward-compatible fixes.

4. Create and push an annotated tag. Replace `vX.Y.Z` with the selected version:

   ```fish
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

5. Verify that the `release-version` workflow passes and that GitHub contains the new release. Do not move or reuse a published tag. Publish a new patch version to correct a release.
