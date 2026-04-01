class Clockscale < Formula
  desc "Terminal UI for viewing multiple timezones in a 24-hour grid"
  homepage "https://github.com/dphase/clockscale-go"
  url "https://github.com/dphase/clockscale-go/archive/refs/tags/v1.2.2.tar.gz"
  sha256 ""
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "."
  end

  test do
    assert_match "clockscale", bin/"clockscale"
  end
end
