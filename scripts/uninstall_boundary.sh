#!/bin/bash
set -e

info() {
  echo '[INFO] ->' "$@"
}

fatal() {
  echo '[ERROR] ->' "$@"
  exit 1
}

verify_system() {
  if ! [ -d /run/systemd ]; then
    fatal 'Can not find systemd to use as a process supervisor for Boundary'
  fi
}

setup_env() {
  SUDO=sudo
  if [ "$(id -u)" -eq 0 ]; then
    SUDO=
  fi

  CONSUL_DATA_DIR=/opt/boundary
  CONSUL_CONFIG_DIR=/etc/boundary.d
  CONSUL_SERVICE_FILE=/etc/systemd/system/boundary.service
  BIN_DIR=/usr/local/bin
}

stop_and_disable_service() {
  info "Stopping and disabling Boundary systemd service"
  $SUDO systemctl stop boundary
  $SUDO systemctl disable boundary
  $SUDO systemctl daemon-reload
}

clean_up() {
  info "Removing Boundary installation"
  $SUDO rm -rf $CONSUL_CONFIG_DIR
  $SUDO rm -rf $CONSUL_DATA_DIR
  $SUDO rm -rf $CONSUL_SERVICE_FILE
  $SUDO rm -rf $BIN_DIR/boundary
}

verify_system
setup_env
stop_and_disable_service
clean_up