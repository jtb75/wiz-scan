**License and Support:**
By using this software and associated documentation files (the “Software”) you hereby agree and understand that:
  - The use of the Software is free of charge and may only be used by Wiz customers for its internal purposes.
  - The Software should not be distributed to third parties.
  - The Software is not part of Wiz’s Services and is not subject to your company’s services agreement with Wiz.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL WIZ BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE USE OF THIS SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.


**Usage of wiz-scan:**

-save
> Set to true to save the configuration

-scanCloudType string
> Possible Values: "AWS", "Azure", "GCP", "OCI", "Alibaba Cloud",
> "Linode", "vSphere"

-scanProviderId string
> External ID for the VM Resource

-scanSubscriptionId string
> Subscription ID (not the name) containing the VM to be scanned

-wizAuthUrl string
> https://auth.app.wiz.io/oauth/token

-wizClientId string
> Service Account ID

-wizClientSecret string
> Service Account Secret

-wizQueryUrl string
> API Endpoint Obtained from Console

-install
> Set to install for recurring scans

-uninstall
> Uninstall from recurring scans

**Examples**

Run from Command Line:

    wiz-scan \
	    -wizClientId SERVICE-ID \
	    -wizClientSecret SERVICE-SECRET \
	    -wizQueryUrl https://api.DC.app.wiz.io/graphql \
	    -wizAuthUrl https://auth.app.wiz.io/oauth/token \
	    -scanCloudType AWS \
	    -scanSubscriptionId 180310087040 \
	    -scanProviderId i-abcd1234ef295685b7b \

Save Config to File:	    

    wiz-scan \
	    -wizClientId SERVICE-ID \
	    -wizClientSecret SERVICE-SECRET \
	    -wizQueryUrl https://api.DC.app.wiz.io/graphql \
	    -wizAuthUrl https://auth.app.wiz.io/oauth/token \
	    -scanCloudType AWS \
	    -scanSubscriptionId 180310087040 \
	    -scanProviderId i-abcd1234ef295685b7b \
	    -save

Run with Saved Config:

    wiz-scan

Install to Task Scheduler or Cron:

    wiz-scan \
        -wizClientId SERVICE-ID \
        -wizClientSecret SERVICE-SECRET \
        -wizQueryUrl https://api.DC.app.wiz.io/graphql \
        -wizAuthUrl https://auth.app.wiz.io/oauth/token \
        -scanCloudType AWS \
        -scanSubscriptionId 180310087040 \
        -scanProviderId i-abcd1234ef295685b7b \
        -install

Uninstall:

    wiz-scan -uninstall
