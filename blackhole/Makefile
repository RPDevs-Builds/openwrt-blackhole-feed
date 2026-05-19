include $(TOPDIR)/rules.mk

PKG_NAME:=blackhole
PKG_VERSION:=0.1.0
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=git
PKG_SOURCE_URL:=https://github.com/IamRPDev/blackhole-server.git
PKG_SOURCE_VERSION:=main

PKG_MAINTAINER:=Gemini CLI <gemini-cli@example.com>
PKG_LICENSE:=MIT
PKG_LICENSE_FILES:=LICENSE

PKG_BUILD_DEPENDS:=golang/host
include $(INCLUDE_DIR)/package.mk
include $(INCLUDE_DIR)/golang-package.mk

define Package/blackhole
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=Blackhole Webserver for Request Tracking
  URL:=https://github.com/IamRPDev/blackhole-server
  DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/blackhole/description
  A lightweight Go webserver that logs all requests and mirrors their paths to the filesystem.
  Returns a 1x1 transparent tracking GIF for all requests.
endef

GO_PKG:=github.com/IamRPDev/blackhole-server
GO_PKG_BUILD_PKG:=github.com/IamRPDev/blackhole-server/cmd/blackhole

define Package/blackhole/install
	$(call GoPackage/Package/Install/Bin,$(1))
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) $(PKG_BUILD_DIR)/openwrt/blackhole.config $(1)/etc/config/blackhole
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/openwrt/blackhole.init $(1)/etc/init.d/blackhole
endef

$(eval $(call GoPackage,blackhole))
$(eval $(call BuildPackage,blackhole))
