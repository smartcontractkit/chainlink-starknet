#[starknet::component]
mod EnumerableSetComponent {
    use core::array::ArrayTrait;

    // set is 1-indexed, not 0-indexed
    #[storage]
    pub struct Storage {
        // access index by value
        // set_id -> item_value -> item_index
        // note: item_index is +1 because 0 means item is not in set
        pub _indexes: LegacyMap::<(u256, u256), u256>,
        // access value by index
        // set_id -> item_id -> item_value
        // note: item_index is +1 because 0 means item is not in set
        // note: _values.read(set_id, item_id) == 0, is only valid iff item_id <= _length.read(set_id)
        pub _values: LegacyMap::<(u256, u256), u256>,
        // set_id -> size of set
        pub _length: LegacyMap<u256, u256>
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {}


    #[generate_trait]
    pub impl InternalImpl<
        TContractState, +HasComponent<TContractState>
    > of InternalTrait<TContractState> {
        fn add(ref self: ComponentState<TContractState>, set_id: u256, value: u256) -> bool {
            if !self.contains(set_id, value) {
                // The value is stored at _length-1, but we add 1 to all indexes
                let index = self._length.read(set_id) + 1;
                self._indexes.write((set_id, value), index);
                self._values.write((set_id, index), value);
                self._length.write(set_id, index);
                true
            } else {
                false
            }
        }

        // swap target value with the last value in the set
        fn remove(
            ref self: ComponentState<TContractState>, set_id: u256, target_value: u256
        ) -> bool {
            let target_index = self._indexes.read((set_id, target_value));
            if target_index == 0 {
                false
            } else {
                let last_index = self._length.read(set_id);
                let last_value = self._values.read((set_id, last_index));

                // if we are NOT trying to remove the last element
                // update the last element mappings
                if last_index != target_index {
                    self._indexes.write((set_id, last_value), target_index);
                    self._values.write((set_id, target_index), last_value);
                }

                // if we are removing the last element both target value and last_index 
                // refer to the same item. 
                self._indexes.write((set_id, target_value), 0);
                self._values.write((set_id, last_index), 0);

                // decrement length of set by 1
                self._length.write(set_id, last_index - 1);

                true
            }
        }

        fn contains(self: @ComponentState<TContractState>, set_id: u256, value: u256) -> bool {
            self._indexes.read((set_id, value)) != 0
        }

        fn length(self: @ComponentState<TContractState>, set_id: u256) -> u256 {
            self._length.read(set_id)
        }

        fn at(self: @ComponentState<TContractState>, set_id: u256, index: u256) -> u256 {
            assert(index != 0, 'set is 1-indexed');
            assert(index <= self._length.read(set_id), 'index out of bounds');
            self._values.read((set_id, index))
        }

        fn values(self: @ComponentState<TContractState>, set_id: u256) -> Array<u256> {
            let len = self.length(set_id);

            let mut result: Array<u256> = ArrayTrait::new();

            let mut i = 1;
            while i <= len {
                result.append(self.at(set_id, i));
                i += 1;
            };

            result
        }
    }
}
