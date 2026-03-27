# typed: false
# frozen_string_literal: true

cask "billy" do
  version "0.1.0"
  sha256 "REPLACE_WITH_SHA256"

  url "https://github.com/jd4rider/billy-wails/releases/download/v#{version}/billy-macos-universal.tar.gz"
  name "Billy"
  desc "Local AI coding assistant desktop app with bundled Billy CLI"
  homepage "https://billysh.online"

  app "Billy.app"
  binary "billy"
end
