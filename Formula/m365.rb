class M365 < Formula
  desc "Microsoft 365 CLI for one signed-in user"
  homepage "https://github.com/jacobhuemmer/m365"
  license "MIT"
  head "https://github.com/jacobhuemmer/m365.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/m365"
  end

  test do
    assert_match "m365", shell_output("#{bin}/m365 --help")
  end
end
