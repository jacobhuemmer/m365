class M365 < Formula
  desc "Microsoft 365 CLI for one signed-in user"
  homepage "https://github.com/jacobhuemmer/m365"
  license "MIT"
  url "https://github.com/jacobhuemmer/m365/archive/refs/tags/v0.0.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  head "https://github.com/jacobhuemmer/m365.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[-s -w -X github.com/masonhuemmer/m365/internal/version.Version=#{version}]
    system "go", "build", *std_go_args(ldflags: ldflags.join(" ")), "./cmd/m365"
  end

  test do
    assert_match "m365", shell_output("#{bin}/m365 --help")
    assert_match version.to_s, shell_output("#{bin}/m365 --version")
  end
end
