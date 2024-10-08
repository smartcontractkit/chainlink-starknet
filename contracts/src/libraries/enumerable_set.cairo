#[starknet::component]
mod EnumerableSetComponent {
    use core::array::ArrayTrait;
    use starknet::{
        StorageAddress,
        storage::{
            Map, StoragePointerReadAccess, StoragePointerWriteAccess, StorageMapReadAccess,
            StorageMapWriteAccess, StoragePathEntry
        }
    };

    // set is 1-indexed, not 0-indexed
    #[storage]
    pub struct Storage {
        // access index by value
        // set_id -> item_value -> item_index
        // note: item_index is +1 because 0 means item is not in set
        pub _indexes: Map<felt252, Map<felt252, usize>>,
        // access value by index
        // set_id -> item_index -> item_value
        // note: item_index is +1 because 0 means item is not in set
        // note: _values.read(set_id, item_id) == 0, is only valid iff item_id <=
        // _length.read(set_id)
        pub _values: Map<felt252, Map<usize, felt252>>,
        // set_id -> size of set
        pub _length: Map<felt252, usize>
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {}


    #[generate_trait]
    pub impl InternalImpl<
        TContractState, +HasComponent<TContractState>
    > of InternalTrait<TContractState> {
        fn add(ref self: ComponentState<TContractState>, set_id: felt252, value: felt252) -> bool {
            if !self.contains(set_id, value) {
                // The value is stored at _length-1, but we add 1 to all indexes
                let index = self._length.read(set_id) + 1;
                self._indexes.entry(set_id).entry(value).write(index);
                self._values.entry(set_id).entry(index).write(value);
                self._length.entry(set_id).write(index);
                true
            } else {
                false
            }
        }

        // swap target value with the last value in the set
        fn remove(
            ref self: ComponentState<TContractState>, set_id: felt252, target_value: felt252
        ) -> bool {
            let target_index = self._indexes.entry(set_id).entry(target_value).read();
            if target_index == 0 {
                false
            } else {
                let last_index = self._length.entry(set_id).read();
                let last_value = self._values.entry(set_id).entry(last_index).read();

                // if we are NOT trying to remove the last element
                // update the last element mappings
                if last_index != target_index {
                    self._indexes.entry(set_id).entry(last_value).write(target_index);
                    self._values.entry(set_id).entry(target_index).write(last_value);
                }

                // if we are removing the last element both target value and last_index
                // refer to the same item.
                self._indexes.entry(set_id).entry(target_value).write(0);
                self._values.entry(set_id).entry(last_index).write(0);

                // decrement length of set by 1
                self._length.entry(set_id).write(last_index - 1);

                true
            }
        }

        fn contains(
            self: @ComponentState<TContractState>, set_id: felt252, value: felt252
        ) -> bool {
            self._indexes.entry(set_id).entry(value).read() != 0
        }

        fn length(self: @ComponentState<TContractState>, set_id: felt252) -> usize {
            self._length.entry(set_id).read()
        }

        fn at(self: @ComponentState<TContractState>, set_id: felt252, index: usize) -> felt252 {
            assert(index != 0, 'set is 1-indexed');
            assert(index <= self._length.entry(set_id).read(), 'index out of bounds');
            self._values.entry(set_id).entry(index).read()
        }

        fn values(self: @ComponentState<TContractState>, set_id: felt252) -> Array<felt252> {
            let len = self.length(set_id);

            let mut result: Array<felt252> = ArrayTrait::new();

            let mut i = 1;
            while i <= len {
                result.append(self.at(set_id, i));
                i += 1;
            };

            result
        }
    }
}
