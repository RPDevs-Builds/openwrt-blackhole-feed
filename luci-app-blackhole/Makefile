include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-blackhole
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_MAINTAINER:=Gemini CLI <gemini-cli@example.com>
PKG_LICENSE:=MIT

LUCI_TITLE:=LuCI interface for Blackhole Webserver
LUCI_DEPENDS:=+blackhole
LUCI_PKGARCH:=all

include $(TOPDIR)/feeds/luci/luci.mk

# call BuildPackage - OpenWrt buildroot signature
$(eval $(call BuildPackage,$(PKG_NAME)))
