class SslPairgen < Formula
  desc "A utility for creating a browser .p12 file and a CA for installation in a Kubernetes cluster"
  homepage "https://github.com/lnxjedi/ssl-pairgen"
  version "0.2.0"

  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/lnxjedi/ssl-pairgen/releases/download/v#{version}/ssl-pairgen-darwin-arm64.tar.gz"
      sha256 "a5ea9f4afa66db69764f4d6c89c336e49607dd72ec25121036989d803e6cb70d"
    else
      url "https://github.com/lnxjedi/ssl-pairgen/releases/download/v#{version}/ssl-pairgen-darwin-amd64.tar.gz"
      sha256 "a338f00743518cbb9f774160bbc76abca7181bda1a1968dd1f7bced98adea200"
    end
  end

  def install
    bin.install "ssl-pairgen"
  end
end
