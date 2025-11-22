class Srps < Formula
  desc "System Resource Protection Script (Ananicy-cpp + EarlyOOM + helpers)"
  homepage "https://github.com/Dicklesworthstone/system_resource_protection_script"
  url "https://github.com/Dicklesworthstone/system_resource_protection_script/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "1c377d8fdfe3dbbb5a6270cbd19d9ea8813b59c1b585c2cd9c681db4806de9ee"
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

      Note: installer makes system-wide changes (ananicy rules, EarlyOOM config, sysctl).
    EOS
  end

  test do
    assert_predicate libexec/"install.sh", :exist?
  end
end
