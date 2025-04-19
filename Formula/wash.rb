class Wash < Formula
  desc "A development assistant that helps track errors, decisions, and project state"
  homepage "https://github.com/brinleekidd/wash-cli"
  url "https://github.com/brinleekidd/wash-cli/releases/download/v0.1.0/wash_Darwin_x86_64.tar.gz"
  sha256 "PLACEHOLDER_SHA256"  # This will be updated after the release is created
  license "MIT"

  depends_on "go" => :build

  def install
    bin.install "wash"
  end

  test do
    system "#{bin}/wash", "version"
  end
end 