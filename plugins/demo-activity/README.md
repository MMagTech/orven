# Demo Activity

A sample Orven plugin. It reports pretend media-server activity from
bundled fixture data so you can see a complete briefing without
connecting anything real. It makes no network requests.

Use the **Scenario to simulate** setting to preview how Orven presents a
quiet day, an unreachable source, or a credential problem.

## Test

```bash
python -m unittest discover -s tests
```

This plugin is also the reference for writing your own — see
`docs/PLUGIN_SDK.md` in the main repository.
