## Intergral Grafana

This is a modified version of Grafana used by Intergral on FusionReactor Cloud platform.

## Update process

The update process for Grafana can be complex. If we are updating a new major oor minor version e.g. there is a new
maintenance branch for the version like v11.2.x. Then we have to reapply our changes mostly manually. If it is a micro
update e.g. v11.2.3, and we have a v11.2.x branch then we can just do a merge.

## Major update path

To update this project when a new version of [Grafana](https://github.com/grafana/grafana) is released follow these
steps:

1. CheckOut this project locally

```bash
  git checkout git@github.com:intergral/grafana.git
```

2. Add a new remote for the upstream [grafana/grafana](https://github.com/grafana/grafana) project

```bash
 git remote add grafana git@github.com:grafana/grafana.git
```

3. Fetch the upstream tags etc

```bash
git fetch grafana
```

4. Create a new branch for the grafana maintenance branch

```bash
git checkout -b v11.2.x grafana/v11.2.x
```

5. Push branch to act as our maintenance branch - this is where we will merge into and tag from to release our changes

```bash
git push -u orign v11.2.x
```

6. Now create a new branch for us to work from
```bash
git branch -b update_v11_2_x origin/v11.2.x
```

7. Now we have to apply our changes to this branch. To do this the easiest way to is to create a patch from `our_changes` branch and apply then to this branch.
   8. Create a patch
   ```bash
    # Checkout the branch with our changes on it
    git checkout origin/our_changes
    # Create a patch from our first commit to HEAD
    git format-patch cb1b5eae81f089fe039495895da8c298d665d618..HEAD --stdout > our_changes.patch
    # Go back to the branch we want to apply the changes to
    git checkout update_v11_2_x
   ```
   9. Apply the patch
   ```bash
   # This will apply as many of our changes as it could. It will create a .rej file for any change it could not apply
   git apply our_changes.patch --reject
   ```
   10. Now review the output and ensure that all the changes have been applied dealing with any .rej files
   11. Ensure that all the original workflows are removed and that only our workflow files are included
   12. Push the changes and create an PR from your new branch to the target maintenance branch e.g. `update_v11_2_x` -> `v11.2.x`
13. Review PR, test and make any additional changes
14. Once happy merge and tag from the maintenance branch
15. We now need to ensure that any additional changes are copied to the `our_changes` branch. This should NOT use merge or rebase.
   

## Known changes

Below are some changes that we have made for this project vs [grafana/grafana](https://github.com/grafana/grafana)

### Build

The build workflows have been re worked to support our needs when updating do not accept any changes from upstream.

## License

Grafana is distributed under [AGPL-3.0-only](LICENSE). For Apache-2.0 exceptions,
see [LICENSING.md](https://github.com/grafana/grafana/blob/HEAD/LICENSING.md).
