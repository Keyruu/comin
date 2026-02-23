{ self }:
{
  config,
  pkgs,
  lib,
  ...
}:
let
  cfg = config;
  cominConfigLib = import ./comin-config.nix { inherit config pkgs lib; };
  inherit (cominConfigLib) cominConfig cominConfigYaml;

  inherit (pkgs.stdenv.hostPlatform) system;
  inherit (cfg.services.comin) package;
in
{
  imports = [ ./module-options.nix ];
  config = lib.mkIf cfg.services.comin.enable {
    assertions = [
      {
        assertion = package != null;
        message = "`services.comin.package` cannot be null.";
      }
      # If the package is null and our `system` isn't supported by the Flake, it's probably safe to show this error message
      {
        assertion = package == null -> lib.elem system (lib.attrNames self.packages);
        message = "comin: ${system} is not supported by the Flake.";
      }
    ];

    systemd.user.services.comin-desktop = lib.mkIf cfg.services.comin.desktop.enable {
      wantedBy = [ "graphical-session.target" ];
      path = [ pkgs.libnotify ];
      serviceConfig = {
        ExecStart = ''${lib.getExe package} desktop --title "${cfg.services.comin.desktop.title}"'';
      };
    };

    environment.systemPackages = [ package ];
    networking.firewall.allowedTCPPorts = lib.optional cfg.services.comin.exporter.openFirewall cfg.services.comin.exporter.port;
    # Use package from overlay first, then Flake package if available
    services.comin.package = lib.mkDefault pkgs.comin or self.packages.${system}.comin or null;
    systemd.services.comin = {
      wantedBy = [ "multi-user.target" ];
      path = [
        config.nix.package
        config.programs.ssh.package
      ];
      # The comin service is restarted by comin itself when it
      # detects the unit file changed.
      restartIfChanged = false;
      serviceConfig = {
        ExecStart =
          (lib.getExe package)
          + (lib.optionalString cfg.services.comin.debug " --debug ")
          + " run "
          + "--config ${cominConfigYaml}";
        Restart = "always";
      };
    }
    // lib.optionalAttrs (cfg.services.comin.sshKeyPath != null) {
      environment = {
        GIT_SSH_COMMAND = "ssh -i ${cfg.services.comin.sshKeyPath} -o UserKnownHostsFile=${cfg.services.comin.sshKnownHostsPath}";
        SSH_KNOWN_HOSTS = cfg.services.comin.sshKnownHostsPath;
        NIX_SSH_KEYS = cfg.services.comin.sshKeyPath;
      };
    };
  };
}
