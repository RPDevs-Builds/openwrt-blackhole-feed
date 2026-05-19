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

# We must include golang-package.mk from the feed, but it has internal relative inclusions
# that assume it's in the main include directory. 
# We'll use the symlink created in the workflow to include it.
-include $(INCLUDE_DIR)/golang-package.mk

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

# Fallback in case golang-package.mk inclusion fails or doesn't define standard macros
ifndef GoPackage/Build/Compile
define Build/Compile
	$(STAGING_DIR_HOST)/bin/go build -o $(PKG_BUILD_DIR)/blackhole $(GO_PKG_BUILD_PKG)
endef
endif

define Package/blackhole/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/blackhole $(1)/usr/bin/blackhole
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) $(PKG_BUILD_DIR)/openwrt/blackhole.config $(1)/etc/config/blackhole
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/openwrt/blackhole.init $(1)/etc/init.d/blackhole
endef

$(eval $(call BuildPackage,blackhole))
