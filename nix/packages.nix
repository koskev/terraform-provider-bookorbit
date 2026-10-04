{ self, ... }: {
  perSystem =
    { pkgs, ... }:
    {
      packages.default =
        pkgs.buildGoModule.override
          {
            go = pkgs.go_1_26;
          }
          {
            pname = "terraform-provider-bookorbit";
            version = "0.1.0";
            src = self;
            vendorHash = "sha256-q9Blhf+SNX+dY74Tm/qYrKFNqRFAzrhg2+vW/NF4JsU=";

            postInstall = ''
              mkdir -p $out/libexec/terraform-providers
              cp $out/bin/bookorbit-provider $out/libexec/terraform-providers/terraform-provider-bookorbit
            '';
          };
    };
}
