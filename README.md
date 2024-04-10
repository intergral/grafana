## Intergral Grafana

This is a modified version of Grafana used by Intergral on FusionReactor Cloud platform.

## Update process

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

4. Create a new branch for the new update

```bash
git checkout -b update_v10_4 tags/v10.4.1
```

5. Push branch to create PR

```bash
git push -u orign updaate_v10_4
```

6. Create PR in Github
7. ReMerge our changes from origin/main - (I recommend using IDE to perform this action)
```bash
git pull origin main --no-rebase
```
![ide_merge.png](ide_merge_screenshot.png)

It will be necessary to manually resolve the conflicts.

The above steps will create a new branch with a PR with all the changes from the old version to the new version. It is
then necessary to check for custom changes on the main branch made by us.

## Known changes

Below are some changes that we have made for this project vs [grafana/grafana](https://github.com/grafana/grafana)

### Build

The build workflows have been re worked to support our needs when updating do not accept any changes from upstream.

## License

Grafana is distributed under [AGPL-3.0-only](LICENSE). For Apache-2.0 exceptions,
see [LICENSING.md](https://github.com/grafana/grafana/blob/HEAD/LICENSING.md).
