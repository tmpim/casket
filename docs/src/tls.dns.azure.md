# tls.dns.azure

<script setup>
import NewInCasket from "./components/NewInCasket.vue";
</script>

Allows you to obtain certificates using DNS records for domains managed with Microsoft Azure. Credentials must be passed
either via environment variables, or directly in the Casketfile.

This provider supports authentication using the **Client Credentials** (Azure AD Application ID and Secret) through
[azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go). The provider is based on the
[libdns/azure](https://github.com/libdns/azure) module.

You will need to create a service principal using [Azure
CLI](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli) or [Azure
Portal](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal), and
assign the **DNS Zone Contributor** role to the service principal for the DNS zones that you want to manage.

## Environment Variables

- `AZURE_TENANT_ID` - Azure tenant ID (Azure Active Directory > Properties > Tenant ID)
- `AZURE_CLIENT_ID` - Azure client ID (Azure Active Directory > App registrations > Your application > Application (client) ID)
- `AZURE_CLIENT_SECRET` - Azure client secret (Azure Active Directory > App registrations > Your application > Certificates & secrets > New client secret)
- `AZURE_SUBSCRIPTION_ID` - Azure subscription ID (DNS Zones > Your Zone > Subscription ID)
- `AZURE_RESOURCE_GROUP` - Azure resource group (DNS Zones > Your Zone > Resource group)

## Syntax

<NewInCasket version="v1.4.0" /> Block configuration syntax:

``` casketfile
tls {
  dns azure {
    tenant_id       TENANT_ID
    client_id       CLIENT_ID
    client_secret   CLIENT_SECRET
    subscription_id SUBSCRIPTION_ID
    resource_group  RESOURCE_GROUP
  }
}
```

- **tenant_id** is your Azure tenant ID (Azure Active Directory > Properties > Tenant ID).
- **client_id** is your Azure client ID (Azure Active Directory > App registrations > Your application > Application (client) ID).
- **client_secret** is your Azure client secret (Azure Active Directory > App registrations > Your application > Certificates & secrets > New client secret).
- **subscription_id** is your Azure subscription ID (DNS Zones > Your Zone > Subscription ID).
- **resource_group** is your Azure resource group (DNS Zones > Your Zone > Resource group).
