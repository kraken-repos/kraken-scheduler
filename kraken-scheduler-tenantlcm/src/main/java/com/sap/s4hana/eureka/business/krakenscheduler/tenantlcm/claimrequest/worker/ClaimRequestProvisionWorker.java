package com.sap.s4hana.eureka.business.krakenscheduler.tenantlcm.claimrequest.worker;

import com.sap.s4hana.eureka.framework.tenant.lcm.provision.AppTenantProvisionWorker;
import com.sap.s4hana.eureka.framework.tenant.lcm.provision.ProvisionBizPayLoad;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

@Component
@Slf4j
public class ClaimRequestProvisionWorker implements AppTenantProvisionWorker {
    @Override
    public void doProvision(ProvisionBizPayLoad provisionBizPayLoad) {
        log.info(
                "Do provision: Company: {}; E-mail: {}",
                provisionBizPayLoad.getCompanyName(),
                provisionBizPayLoad.getAdminEmail());
    }
}
