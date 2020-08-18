## Running tl in Docker

This uses Docker to run tl to verify an existing file from the current working
directory on the host machine.

```
docker run \
	quay.io/transparencylog/tl:latest \
	--mount type=bind,source=$PWD,target=/mnt \
	/tl verify https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.1.tar.xz /mnt/linux-5.1.tar.xz
```

The return code will be forwarded from docker so you can test for non-zero exit
for verification failure.
