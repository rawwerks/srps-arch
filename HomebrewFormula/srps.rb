class Srps < Formula
  desc "System Resource Protection Script (Ananicy-cpp + helpers)"
  homepage "https://github.com/Dicklesworthstone/system_resource_protection_script"
  url "https://github.com/Dicklesworthstone/system_resource_protection_script/archive/refs/tags/v1.1.3.tar.gz"
  sha256 "7310bd4bd13705d49c1cbedfad95cccd39cf5a2ab0d88b07db80ceb7b9fcf048"
  license "MIT"

  head "https://github.com/Dicklesworthstone/system_resource_protection_script.git", branch: "main"

  depends_on "bash"
  depends_on "git"
  depends_on "jq"

  def install
    libexec.install "install.sh", "verify.sh", "README.md"
    bin.install_symlink libexec/"install.sh" => "srps-install"
    bin.install_symlink libexec/"verify.sh" => "srps-verify"
  end

  def caveats
    <<~EOS
      Usage:
        srps-verify latest   # verify release integrity
        srps-install --plan  # dry-run the installer
        srps-install         # run installer (prompts for sudo)

      Note: installer makes system-wide changes (ananicy rules, sysctl).
    EOS
  end

  test do
    assert_predicate libexec/"install.sh", :exist?
  end
end
