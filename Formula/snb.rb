class Snb < Formula
  desc "ShakeAndBake CLI"
  homepage "https://github.com/aemengo/snb"
  version "0.1.0"
  url "https://github.com/aemengo/snb/releases/download/v0.1.0/snb-darwin-0.1.0.tgz"
  sha256 "9a47b694f47c0ae39a98fe6c4ea3bcced5a2ccd6c5fd8a8706cbc57fa0048844"

  depends_on :arch => :x86_64

  def install
    bin.install "snb"
  end

  test do
    system "#{bin}/snb --help"
  end
end
