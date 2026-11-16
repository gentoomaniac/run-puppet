pkgname=run-puppet
pkgver=0.1.8
pkgrel=1
pkgdesc="Wrapper to run puppet in my homelab"
arch=('x86_64')
url="https://github.com/gentoomaniac/run-puppet"
license=('MIT')
options=('!debug')

# Point this to where your GoReleaser uploads the tarball
source=("${pkgname}::${url}/releases/download/v${pkgver}/${pkgname}_${pkgver}_linux-amd64")
sha256sums=('7804f72f379ef53b9a460b84adb68050266f2a51caae45d2df89c1972080df88')

package() {
  # GoReleaser usually puts the binary in the root of the tarball
  install -Dm755 "${srcdir}/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
}
