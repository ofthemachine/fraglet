# Alpine Fraglet Help

This container supports fraglet injection into bash scripts.

## Usage

The fraglet will be injected at the `FRAGLET` marker in `/code/hello-world.sh`.

## Example

```bash
docker run -v /path/to/fraglet:/FRAGLET myimage
```

