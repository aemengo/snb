#!/usr/bin/env bash

set -e

if [[ -z ${1} ]]; then
  echo "[USAGE]: ${0} RELEASE_VERSION"
  exit 1
fi

dir=$(cd `dirname ${0}` && cd .. && pwd)
version=$1

mkdir -p ${dir}/out

GOARCH=amd64 GOOS=darwin go \
  build \
  -o ${dir}/out/snb \
  github.com/aemengo/snb

chmod +x ${dir}/out/snb
tar czvf ${dir}/out/snb-darwin-${version}.tgz -C ${dir}/out snb
shasum=$(shasum -a 256 ${dir}/out/snb-darwin-${version}.tgz | cut -d ' ' -f 1)

cat > ${dir}/Formula/snb.rb <<EOF
class SnbCli < Formula
  desc "ShakeAndBake CLI"
  homepage "https://github.com/aemengo/snb"
  version "${version}"
  url "https://github.com/aemengo/snb/releases/download/v${version}/snb-darwin-${version}.tgz"
  sha256 "${shasum}"

  depends_on :arch => :x86_64

  def install
    bin.install "snb"
  end

  test do
    system "#{bin}/snb --help"
  end
end
EOF

git add ${dir}/Formula
git commit -m "Release: ${version}"
