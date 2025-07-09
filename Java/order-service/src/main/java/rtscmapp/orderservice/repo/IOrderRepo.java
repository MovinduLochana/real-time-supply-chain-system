package rtscmapp.orderservice.repo;

import org.springframework.data.jpa.repository.JpaRepository;
import rtscmapp.orderservice.model.IOrder;

public interface IOrderRepo extends JpaRepository<IOrder, Integer> {

}
