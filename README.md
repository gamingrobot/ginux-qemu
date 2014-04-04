ginux
===
Run Linux in the browser

http://arm.gamingrobot.net

Using your own kernel/busybox (Optional)
===

### Compile kernel
http://www.kernel.org/pub/linux/kernel/v2.6/linux-2.6.33.tar.bz2
```
make ARCH=arm versatile_defconfig
make ARCH=arm menuconfig
	Remove module support
```

```
qemu-system-arm -M versatilepb -m 20M -kernel arch/arm/boot/zImage
```

### Compile Busybox
http://www.busybox.net/downloads/busybox-1.22.1.tar.bz2

```
make ARCH=arm CROSS_COMPILE=arm-linux-gnueabi- defconfig
make ARCH=arm CROSS_COMPILE=arm-linux-gnueabi- menuconfig
	Static version in busybox settings-> build options
make ARCH=arm CROSS_COMPILE=arm-linux-gnueabi- install
cd _install
mkdir proc sys dev etc etc/init.d
vim etc/init.d/rcS
```

```
#!/bin/sh
mount -t proc none /proc
mount -t sysfs none /sys
/sbin/mdev -s
```

```
find . | cpio -o --format=newc > ../../rootfs.img
cd ../../
gzip -c rootfs.img > rootfs.img.gz
```

```
qemu-system-arm -M versatilepb -m 20M -nographic -readconfig qemu.conf
```
