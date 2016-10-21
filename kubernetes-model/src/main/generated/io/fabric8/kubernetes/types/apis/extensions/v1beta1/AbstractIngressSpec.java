package io.fabric8.kubernetes.types.apis.extensions.v1beta1;

import org.immutables.value.Value;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonUnwrapped;

/*
 * IngressSpec describes the Ingress the user wishes to exist.
 */
@Value.Immutable
abstract class AbstractIngressSpec {
  /*
   * A default backend capable of servicing requests that don't match any
   * rule. At least one of 'backend' or 'rules' must be specified. This field
   * is optional to allow the loadbalancer controller or defaulting logic to
   * specify a global default.
   */
  @JsonProperty("backend")
  public abstract io.fabric8.kubernetes.types.apis.extensions.v1beta1.IngressBackend getBackend();

  /*
   * TLS configuration. Currently the Ingress only supports a single TLS
   * port, 443. If multiple members of this list specify different hosts, they
   * will be multiplexed on the same port according to the hostname specified
   * through the SNI TLS extension, if the ingress controller fulfilling the
   * ingress supports SNI.
   */
  @JsonProperty("tls")
  public abstract java.util.List<io.fabric8.kubernetes.types.apis.extensions.v1beta1.IngressTLS> getTls();

  /*
   * A list of host rules used to configure the Ingress. If unspecified, or
   * no rule matches, all traffic is sent to the default backend.
   */
  @JsonProperty("rules")
  public abstract java.util.List<io.fabric8.kubernetes.types.apis.extensions.v1beta1.IngressRule> getRules();

}
