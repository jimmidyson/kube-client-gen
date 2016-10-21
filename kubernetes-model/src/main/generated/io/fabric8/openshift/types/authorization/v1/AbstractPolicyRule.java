package io.fabric8.openshift.types.authorization.v1;

import org.immutables.value.Value;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonUnwrapped;

/*
 * PolicyRule holds information that describes a policy rule, but does not contain information
 * about who the rule applies to or which namespace the rule applies to.
 */
@Value.Immutable
abstract class AbstractPolicyRule {
  /*
   * Verbs is a list of Verbs that apply to ALL the ResourceKinds and AttributeRestrictions contained in this rule.  VerbAll represents all kinds.
   */
  @JsonProperty("verbs")
  public abstract java.util.List<String> getVerbs();

  /*
   * AttributeRestrictions will vary depending on what the Authorizer/AuthorizationAttributeBuilder pair supports.
   * If the Authorizer does not recognize how to handle the AttributeRestrictions, the Authorizer should report an error.
   */
  @JsonProperty("attributeRestrictions")
  public abstract io.fabric8.kubernetes.types.api.HasMetadata getAttributeRestrictions();

  /*
   * APIGroups is the name of the APIGroup that contains the resources.  If this field is empty, then both kubernetes and origin API groups are assumed.
   * That means that if an action is requested against one of the enumerated resources in either the kubernetes or the origin API group, the request
   * will be allowed
   */
  @JsonProperty("apiGroups")
  public abstract java.util.List<String> getApiGroups();

  /*
   * Resources is a list of resources this rule applies to.  ResourceAll represents all resources.
   */
  @JsonProperty("resources")
  public abstract java.util.List<String> getResources();

  /*
   * ResourceNames is an optional white list of names that the rule applies to.  An empty set means that everything is allowed.
   */
  @JsonProperty("resourceNames")
  public abstract java.util.List<String> getResourceNames();

  /*
   * NonResourceURLsSlice is a set of partial urls that a user should have access to.  *s are allowed, but only as the full, final step in the path
   * This name is intentionally different than the internal type so that the DefaultConvert works nicely and because the ordering may be different.
   */
  @JsonProperty("nonResourceURLs")
  public abstract java.util.List<String> getNonResourceURLs();

}
