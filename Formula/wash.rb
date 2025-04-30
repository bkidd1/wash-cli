class Wash < Formula
  desc "A development assistant that helps track errors, decisions, and project state"
  homepage "https://github.com/bkidd1/wash-cli"
  url "https://github.com/bkidd1/wash-cli/archive/refs/tags/v1.0.3.tar.gz"
  sha256 "5ac938306d802ecf485b46f992043ac35f8173b9ee60dceb4641fe47b652d103"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/bkidd1/wash-cli/releases/download/v1.0.3/wash-cli_Darwin_x86_64.tar.gz"
      sha256 "5ac938306d802ecf485b46f992043ac35f8173b9ee60dceb4641fe47b652d103"
    end
    if Hardware::CPU.arm?
      url "https://github.com/bkidd1/wash-cli/releases/download/v1.0.3/wash-cli_Darwin_arm64.tar.gz"
      sha256 "5ac938306d802ecf485b46f992043ac35f8173b9ee60dceb4641fe47b652d103"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/bkidd1/wash-cli/releases/download/v1.0.3/wash-cli_Linux_x86_64.tar.gz"
      sha256 "5ac938306d802ecf485b46f992043ac35f8173b9ee60dceb4641fe47b652d103"
    end
    if Hardware::CPU.arm?
      url "https://github.com/bkidd1/wash-cli/releases/download/v1.0.3/wash-cli_Linux_arm64.tar.gz"
      sha256 "5ac938306d802ecf485b46f992043ac35f8173b9ee60dceb4641fe47b652d103"
    end
  end

  def install
    bin.install "wash"
  end

  test do
    assert_match "v1.0.3", shell_output("#{bin}/wash version")
  end
end 