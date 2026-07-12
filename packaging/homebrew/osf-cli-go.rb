class OsfCliGo < Formula
  desc "Command-line client for the Open Science Framework"
  homepage "https://github.com/edithatogo/osf-cli-go"
  version "0.3.2"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/edithatogo/osf-cli-go/releases/download/v0.3.2/osf-darwin-arm64"
      sha256 "144c3bf9295459f304c26227e32481f68079d606d237d5175a9f9f36b73993b9"
    else
      url "https://github.com/edithatogo/osf-cli-go/releases/download/v0.3.2/osf-darwin-amd64"
      sha256 "2bf172de6520449117df5699bd24400585a829f8e0b7b9fe04d1355dacaa04eb"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/edithatogo/osf-cli-go/releases/download/v0.3.2/osf-linux-arm64"
      sha256 "ff4e78256fae17e3ffa9532614a864a0eff2f49124fabfd553bf24961aba82fa"
    else
      url "https://github.com/edithatogo/osf-cli-go/releases/download/v0.3.2/osf-linux-amd64"
      sha256 "4af8d3819d1b34fc19402ea8355b29875d7defb7d4381d49323d307dc58e65ea"
    end
  end

  def install
    bin.install Dir["osf-*"] => "osf"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/osf --version")
  end
end
